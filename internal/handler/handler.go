package handler

import (
	"fmt"
	"github.com/MrSwed/go-musthave-metrics/internal/constants"
	"github.com/MrSwed/go-musthave-metrics/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

func Handler(s *service.Service) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	//r.Use(middleware.URLFormat)
	//r.Route(fmt.Sprintf(`%s/{metricType:%s|%s}/{metricName}/{metricValue}`,
	//	constants.UpdateRoute, constants.MetricTypeCounter, constants.MetricTypeCounter),
	r.Route(fmt.Sprintf(`%s/{metricType}/{metricName}/{metricValue}`, constants.UpdateRoute),
		UpdateHandler(s))

	log.Fatal(http.ListenAndServe(`:8080`, r))
}

func UpdateHandler(s *service.Service) func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", UpdateMetric(s))
	}
}
