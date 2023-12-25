package handler

import (
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
	r.Route(constants.UpdateRoute+"/gauge", UpdateGaugeHandler(s))
	r.Route(constants.UpdateRoute+"/counter", UpdateCounterHandler(s))

	log.Fatal(http.ListenAndServe(`:8080`, r))
}

func UpdateGaugeHandler(s *service.Service) func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/{metricName}/{metricValue}", UpdateGaugeMetric(s))
	}
}
func UpdateCounterHandler(s *service.Service) func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/{metricName}/{metricValue}", UpdateCounterMetric(s))
	}
}
