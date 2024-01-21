package main

import (
	"os"
	"os/signal"
	"syscall"

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
	r, err := repository.NewRepository(&conf.StorageConfig)
	if err != nil {
		panic(err)
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
