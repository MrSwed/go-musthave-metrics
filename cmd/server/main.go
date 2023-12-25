package main

import (
	"flag"
	"github.com/MrSwed/go-musthave-metrics/internal/handler"
	"github.com/MrSwed/go-musthave-metrics/internal/repository"
	"github.com/MrSwed/go-musthave-metrics/internal/service"
)

var (
	serverAddress = flag.String("a", "localhost:8080", "Provide the address of the metrics collection server (without protocol)")
)

func main() {
	flag.Parse()
	r := repository.NewRepository()
	s := service.NewService(r)
	handler.Handler(*serverAddress, s)
}
