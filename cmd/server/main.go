package main

import (
	"github.com/MrSwed/go-musthave-metrics/internal/handler"
	"github.com/MrSwed/go-musthave-metrics/internal/repository"
	"github.com/MrSwed/go-musthave-metrics/internal/service"
)

func main() {
	r := repository.NewRepository()
	s := service.NewService(r)
	handler.Handler(s)
}
