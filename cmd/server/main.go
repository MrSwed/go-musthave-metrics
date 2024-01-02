package main

import (
	"flag"
	"github.com/MrSwed/go-musthave-metrics/internal/handler"
	"github.com/MrSwed/go-musthave-metrics/internal/repository"
	"github.com/MrSwed/go-musthave-metrics/internal/service"
	"log"
	"os"
)

func main() {
	var serverAddress string
	flag.StringVar(&serverAddress, "a", "localhost:8080", "Provide the address of the metrics collection server (without protocol)")
	flag.Parse()

	if addressEnv := os.Getenv("ADDRESS"); addressEnv != "" {
		serverAddress = addressEnv
	}

	log.Printf(`Started with config:
  serverAddress: %s
`, serverAddress)

	r := repository.NewRepository()
	s := service.NewService(r)
	handler.Handler(serverAddress, s)
}
