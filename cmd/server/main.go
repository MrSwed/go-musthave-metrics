package main

import (
	"flag"
	"github.com/MrSwed/go-musthave-metrics/internal/handler"
	"github.com/MrSwed/go-musthave-metrics/internal/repository"
	"github.com/MrSwed/go-musthave-metrics/internal/service"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var serverAddress string
	flag.StringVar(&serverAddress, "a", "localhost:8080", "Provide the address of the metrics collection server (without protocol)")
	flag.Parse()

	if addressEnv := os.Getenv("ADDRESS"); addressEnv != "" {
		serverAddress = addressEnv
	}

	r := repository.NewRepository()
	s := service.NewService(r)
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	logger.Info("Start server", zap.String("serverAddress", serverAddress))

	defer func() {
		logger.Info("Server stopped")
		_ = logger.Sync()
	}()

	go func() {
		if err := handler.NewHandler(s, logger).RunServer(serverAddress); err != nil {
			logger.Error("Can not start server", zap.Error(err))
			os.Exit(1)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	_ = <-c
}
