package main

import (
	"flag"
	"github.com/MrSwed/go-musthave-metrics/internal/handler"
	"github.com/MrSwed/go-musthave-metrics/internal/repository"
	"github.com/MrSwed/go-musthave-metrics/internal/service"
	"log"
	"os"
)

var (
	serverAddress = flag.String("a", "localhost:8080", "Provide the address of the metrics collection server (without protocol)")
)

func getEnv() {
	addressEnv := os.Getenv("ADDRESS")
	if addressEnv != "" {
		*serverAddress = addressEnv
	}
}

func main() {
	flag.Parse()
	getEnv()
	log.Printf(`Started with config:
  serverAddress: %s
`, *serverAddress)

	r := repository.NewRepository()
	s := service.NewService(r)
	handler.Handler(*serverAddress, s)
}
