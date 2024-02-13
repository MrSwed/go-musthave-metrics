package main

import (
	"context"
	"errors"
	"github.com/MrSwed/go-musthave-metrics/internal/constant"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/MrSwed/go-musthave-metrics/internal/config"
	"github.com/MrSwed/go-musthave-metrics/internal/handler"
	myMigrate "github.com/MrSwed/go-musthave-metrics/internal/migrate"
	"github.com/MrSwed/go-musthave-metrics/internal/repository"
	"github.com/MrSwed/go-musthave-metrics/internal/service"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	var wg sync.WaitGroup
	conf := config.NewConfig().Init()

	serverCtx, serverStop := context.WithCancel(context.Background())
	serverFileCtx, serverFileStop := context.WithCancel(context.Background())
	serverDBCtx, serverDBStop := context.WithCancel(context.Background())

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	logger.Info("Start server", zap.Any("Config", conf))

	var (
		db         *sqlx.DB
		isNewStore = true
	)
	if len(conf.DatabaseDSN) > 0 {
		if db, err = sqlx.Connect("postgres", conf.DatabaseDSN); err != nil {
			logger.Fatal("cannot connect db", zap.Error(err))
		}
		isNewStore = false
		logger.Info("DB connected")
		versions, errM := myMigrate.Migrate(db.DB)
		switch {
		case errors.Is(errM, migrate.ErrNoChange):
			logger.Info("DB migrate: ", zap.Any("info", errM), zap.Any("versions", versions))
		case errM == nil:
			logger.Info("DB migrate: new applied ", zap.Any("versions", versions))
			isNewStore = versions[0] == 0
		default:
			logger.Fatal("DB migrate: ", zap.Any("versions", versions), zap.Error(errM))
		}
	}

	r := repository.NewRepository(&conf.StorageConfig, db)
	s := service.NewService(r, &conf.StorageConfig)
	h := handler.NewHandler(s, logger)

	if conf.FileStoragePath != "" {
		if conf.StorageRestore {
			if isNewStore {
				if n, err := s.RestoreFromFile(); err != nil {
					logger.Error("File storage restore", zap.Error(err))
				} else {
					logger.Info("File storage restored success", zap.Any("records", n))
				}
			} else {
				logger.Info("Storage not restored - it is not new db store used")
			}
		}
		if conf.FileStoreInterval > 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-time.After(time.Duration(conf.FileStoreInterval) * time.Second):
						if n, err := s.SaveToFile(); err != nil {
							logger.Error("Storage save", zap.Error(err))
						} else {
							logger.Info("Storage saved", zap.Any("records", n))
						}
					case <-serverCtx.Done():
						logger.Info("FileStoreInterval finished")
						return
					}
				}
			}()
		}
	}

	server := &http.Server{Addr: conf.ServerAddress, Handler: h.Handler()}

	exitSig := make(chan os.Signal, 1)
	signal.Notify(exitSig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-exitSig

		shutdownCtx, shutdownStopForce := context.WithTimeout(serverCtx, constant.ServerShutdownTimeout*time.Second)
		defer shutdownStopForce()
		go func() {
			<-shutdownCtx.Done()
			if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
				logger.Error("graceful shutdown timed out.. forcing exit.", zap.Any("timeout", constant.ServerShutdownTimeout))
			}
		}()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown", zap.Error(err))
		}

		serverStop()
	}()

	go func() {
		<-serverCtx.Done()
		if conf.FileStoragePath != "" {
			shutdownFileCtx, shutdownFileStopForce := context.WithTimeout(serverFileCtx, constant.ServerShutdownTimeout*time.Second)
			defer shutdownFileStopForce()
			var (
				complete = make(chan struct{}, 1)
			)
			go func() {
				if n, err := s.SaveToFile(); err != nil {
					logger.Error("Storage save", zap.Error(err))
				} else {
					logger.Info("Storage saved", zap.Any("records", n))
				}
				complete <- struct{}{}
			}()

			select {
			case <-complete:
				break
			case <-shutdownFileCtx.Done():
				logger.Error("FileSave graceful shutdown timed out.. forcing exit.", zap.Any("timeout", constant.ServerShutdownTimeout))
				return
			}
		}
		serverFileStop()
	}()

	go func() {
		<-serverFileCtx.Done()
		if db != nil {
			shutdownDBCtx, shutdownDBStopForce := context.WithTimeout(serverDBCtx, constant.ServerShutdownTimeout*time.Second)
			defer shutdownDBStopForce()
			var (
				complete = make(chan struct{}, 1)
			)
			go func() {
				if err = db.Close(); err != nil {
					logger.Error("DB close", zap.Error(err))
				} else {
					logger.Info("Db Closed")
				}
				complete <- struct{}{}
			}()

			select {
			case <-complete:
				break
			case <-shutdownDBCtx.Done():
				logger.Error("DB graceful shutdown timed out.. forcing exit.", zap.Any("timeout", constant.ServerShutdownTimeout))
				return
			}
		}

		serverDBStop()
	}()

	logger.Info("Start web app")
	if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("Start server", zap.Error(err))
		serverStop()
	}

	<-serverDBCtx.Done()

	// wait FileStoreInterval
	wg.Wait()
	logger.Info("Server stopped")

	_ = logger.Sync()
}
