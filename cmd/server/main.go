package main

import (
	"context"
	"errors"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/MrSwed/go-musthave-metrics/internal/server/closer"
	"github.com/MrSwed/go-musthave-metrics/internal/server/config"
	"github.com/MrSwed/go-musthave-metrics/internal/server/constant"
	"github.com/MrSwed/go-musthave-metrics/internal/server/handler"
	myMigrate "github.com/MrSwed/go-musthave-metrics/internal/server/migrate"
	"github.com/MrSwed/go-musthave-metrics/internal/server/repository"
	"github.com/MrSwed/go-musthave-metrics/internal/server/service"
	"github.com/go-chi/chi/v5"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	runServer(ctx)
}

func runServer(ctx context.Context) {
	var wg sync.WaitGroup
	conf := config.NewConfig().Init()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	logger.Info("Start server", zap.Any("Config", conf))

	var (
		db         *sqlx.DB
		isNewStore = true
		sCloser    = &closer.Closer{}
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
	h := handler.NewHandler(chi.NewRouter(), s, &conf.WEB, logger)

	if conf.FileStoragePath != "" {
		if conf.StorageRestore {
			if isNewStore {
				if n, er := s.RestoreFromFile(ctx); er != nil {
					logger.Error("File storage restore", zap.Error(er))
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
						if n, er := s.SaveToFile(ctx); er != nil {
							logger.Error("Storage save", zap.Error(er))
						} else {
							logger.Info("Storage saved", zap.Any("records", n))
						}
					case <-ctx.Done():
						logger.Info("Store save on interval finished")
						return
					}
				}
			}()
		}
	}

	server := &http.Server{Addr: conf.ServerAddress, Handler: h.Handler()}
	lockDBCLose := make(chan struct{})

	sCloser.Add("WEB", server.Shutdown)

	if conf.FileStoragePath != "" && conf.FileStoreInterval == 0 {
		sCloser.Add("Storage save", func(ctx context.Context) (err error) {
			defer close(lockDBCLose)
			var n int64
			if n, err = s.SaveToFile(ctx); err == nil {
				logger.Info("Storage saved", zap.Any("records", n))
			}
			return
		})
	} else {
		close(lockDBCLose)
	}

	if db != nil {
		sCloser.Add("DB Close", func(ctx context.Context) (err error) {
			<-lockDBCLose
			if err = db.Close(); err == nil {
				logger.Info("Db Closed")
			}
			return
		})
	}

	go func() {
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Start server", zap.Error(err))
		}
	}()

	logger.Info("Server started")

	<-ctx.Done()

	logger.Info("Shutting down server gracefully")

	// wait FileStoreInterval
	wg.Wait()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), constant.ServerShutdownTimeout*time.Second)
	defer cancel()

	if err = sCloser.Close(shutdownCtx); err != nil {
		logger.Error("Shutdown", zap.Error(err), zap.Any("timeout: ", constant.ServerShutdownTimeout))
	}

	logger.Info("Server stopped")

	_ = logger.Sync()
}
