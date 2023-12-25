package handler

import (
	"github.com/MrSwed/go-musthave-metrics/internal/constants"
	"github.com/MrSwed/go-musthave-metrics/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

func Handler(addr string, s *service.Service) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Route("/", GetListValuesHandler(s))

	//r.Use(middleware.URLFormat)
	//r.Route(fmt.Sprintf(`%s/{metricType:%s|%s}/{metricName}/{metricValue}`,
	//	constants.UpdateRoute, constants.MetricTypeCounter, constants.MetricTypeCounter),
	r.Route(constants.UpdateRoute+"/{metricType}/{metricName}/{metricValue}", UpdateHandler(s))

	r.Route(constants.ValueRoute+"/{metricType}/{metricName}", GetValueHandler(s))

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
