package main

import (
	"context"
	"errors"
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

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	logger.Info("Start server", zap.Any("Config", conf))

	var (
		db      *sqlx.DB
		isNewDB = true
	)
	if len(conf.DatabaseDSN) > 0 {
		if db, err = sqlx.Connect("postgres", conf.DatabaseDSN); err != nil {
			logger.Fatal("cannot connect db", zap.Error(err))
		}
		logger.Info("DB connected")
		versions, errM := myMigrate.Migrate(db.DB)
		switch {
		case errors.Is(errM, migrate.ErrNoChange):
			logger.Info("DB migrate: ", zap.Any("info", errM), zap.Any("versions", versions))
		case errM == nil:
			logger.Info("DB migrate: new applied ", zap.Any("versions", versions))
		default:
			logger.Fatal("DB migrate: ", zap.Any("versions", versions), zap.Error(errM))
		}
		isNewDB = versions[0] == 0

	}

	r := repository.NewRepository(&conf.StorageConfig, db)
	s := service.NewService(r, &conf.StorageConfig)
	h := handler.NewHandler(s, logger)

	if conf.FileStoragePath != "" && isNewDB {
		if conf.StorageRestore {
			if err := s.RestoreFromFile(); err != nil {
				logger.Error("Storage restore", zap.Error(err))
			} else {
				logger.Info("Storage restored")
			}
		}
		if conf.StoreInterval > 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-time.After(time.Duration(conf.StoreInterval) * time.Second):
						if err := s.SaveToFile(); err != nil {
							logger.Error("Storage save", zap.Error(err))
						}
					case <-serverCtx.Done():
						logger.Info("StoreInterval finished")
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

		shutdownCtx, shutdownStopForce := context.WithTimeout(serverCtx, config.ServerShutdownTimeout*time.Second)
		defer shutdownStopForce()
		go func() {
			<-shutdownCtx.Done()
			if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
				logger.Error("graceful shutdown timed out.. forcing exit.", zap.Any("timeout", config.ServerShutdownTimeout))
			}
		}()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown", zap.Error(err))
		}
		serverStop()
	}()

	logger.Info("Start web app")
	if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("Start server", zap.Error(err))
		serverStop()
	}

	<-serverCtx.Done()

	// wait StoreInterval
	wg.Wait()
	logger.Info("Server stopped")

	if conf.FileStoragePath != "" {
		if err := s.SaveToFile(); err != nil {
			logger.Error("Storage save", zap.Error(err))
		} else {
			logger.Info("Storage saved")
		}
	}

	if db != nil {
		if err = db.Close(); err != nil {
			logger.Error("DB close", zap.Error(err))
		} else {
			logger.Info("Db Closed")
		}
	}
	_ = logger.Sync()
}
