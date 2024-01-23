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
	"github.com/MrSwed/go-musthave-metrics/internal/repository"
	"github.com/MrSwed/go-musthave-metrics/internal/service"

	"go.uber.org/zap"
)

func main() {
	var wg sync.WaitGroup
	conf := config.NewConfig(true)

	serverCtx, serverStop := context.WithCancel(context.Background())

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	r := repository.NewRepository(&conf.StorageConfig)
	s := service.NewService(r)
	h := handler.NewHandler(s, logger)

	logger.Info("Start server", zap.Any("Config", conf))

	if conf.FileStoragePath != "" {
		if conf.StorageRestore {
			if err := r.Restore(); err != nil {
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
						if err := r.Save(); err != nil {
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

		shutdownCtx, _ := context.WithTimeout(serverCtx, config.ServerShutdownTimeout*time.Second)
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

	if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("shutdown", zap.Error(err))
	}

	<-serverCtx.Done()

	// wait StoreInterval
	wg.Wait()

	if conf.FileStoragePath != "" {
		if err := r.Save(); err != nil {
			logger.Error("Storage save", zap.Error(err))
		} else {
			logger.Info("Storage saved")
		}
	}
	logger.Info("Server stopped")
	_ = logger.Sync()
}
