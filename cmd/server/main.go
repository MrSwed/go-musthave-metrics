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

	r := repository.NewRepository()
	s := service.NewService(r)
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	logger.Info("Start server", zap.Any("Config", conf))

	defer func() {
		logger.Info("Server stopped")
		_ = logger.Sync()
	}()

	go func() {
		if err := handler.NewHandler(s, conf, logger).RunServer(); err != nil {
			logger.Error("Can not start server", zap.Error(err))
			os.Exit(1)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
