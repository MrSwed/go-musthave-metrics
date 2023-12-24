package handler

import (
	"github.com/MrSwed/go-musthave-metrics/internal/constants"
	"github.com/MrSwed/go-musthave-metrics/internal/service"
	"net/http"
)

func Handler(s *service.Service) {
	mux := http.NewServeMux()
	mux.HandleFunc(constants.UpdateRoute, UpdateMetric(s))

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
