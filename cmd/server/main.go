package main

import (
	"flag"
	"github.com/MrSwed/go-musthave-metrics/internal/handler"
	"github.com/MrSwed/go-musthave-metrics/internal/repository"
	"github.com/MrSwed/go-musthave-metrics/internal/service"
)

var (
	baseURL = flag.String("a", "localhost:8080", "Provide the address of the metrics collection server (without protocol)")
)

func init() {
	flag.Parse()
}

func main() {
	r := repository.NewRepository()
	s := service.NewService(r)
	handler.Handler(*baseURL, s)
}
