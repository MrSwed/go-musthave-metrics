package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MrSwed/go-musthave-metrics/internal/config"
	"github.com/MrSwed/go-musthave-metrics/internal/handler"
	"github.com/MrSwed/go-musthave-metrics/internal/repository"
	"github.com/MrSwed/go-musthave-metrics/internal/service"

	"go.uber.org/zap"
)

func main() {
	conf := config.NewConfig(true)
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	r := repository.NewRepository(&conf.StorageConfig)
	if conf.FileStoragePath != "" {
		if conf.StorageRestore {
			if err := r.Restore(); err != nil {
				logger.Error("Storage restore", zap.Error(err))
			}
		}
		if conf.StoreInterval > 0 {
			go func() {
				for {
					time.Sleep(time.Duration(conf.StoreInterval) * time.Second)
					if err := r.Save(); err != nil {
						logger.Error("Storage save", zap.Error(err))
					}
				}
			}()
		}
	}
	s := service.NewService(r)

	logger.Info("Start server", zap.Any("Config", conf))

	defer func() {
		if err := r.Save(); err != nil {
			logger.Error("Storage save", zap.Error(err))
		} else {
			logger.Info("Storage saved")
		}
		logger.Info("Server stopped")
		_ = logger.Sync()
	}()

	go func() {
		if err := handler.NewHandler(s, &conf.ServerConfig, logger).RunServer(); err != nil {
			logger.Error("Can not start server", zap.Error(err))
			os.Exit(1)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
