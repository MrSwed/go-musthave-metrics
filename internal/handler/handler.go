package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MrSwed/go-musthave-metrics/internal/constants"
	"github.com/MrSwed/go-musthave-metrics/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Handler(addr string, s *service.Service) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Route("/", GetListValuesHandler(s))

	r.Route(fmt.Sprintf("%s/{%s}/{%s}/{%s}",
		constants.UpdateRoute, constants.MetricTypeParam, constants.MetricNameParam, constants.MetricValueParam),
		UpdateHandler(s))

	r.Route(fmt.Sprintf("%s/{%s}/{%s}",
		constants.ValueRoute, constants.MetricTypeParam, constants.MetricNameParam),
		GetValueHandler(s))

	log.Fatal(http.ListenAndServe(addr, r))
}

func UpdateHandler(s *service.Service) func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", UpdateMetric(s))
	}
}

func GetValueHandler(s *service.Service) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", GetMetric(s))
	}
}
func GetListValuesHandler(s *service.Service) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", GetListMetrics(s))
	}
}
