package handler

import (
	"errors"
	"fmt"
	"github.com/MrSwed/go-musthave-metrics/internal/constants"
	myErr "github.com/MrSwed/go-musthave-metrics/internal/errors"
	"github.com/MrSwed/go-musthave-metrics/internal/service"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func GetMetric(s *service.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		action, metricKey := chi.URLParam(r, "metricType"), chi.URLParam(r, "metricName")
		var metricValue string
		switch action {
		case constants.MetricTypeGauge:
			if gauge, err := s.GetGauge(metricKey); err != nil {
				if errors.Is(err, myErr.ErrNotExist) {
					w.WriteHeader(http.StatusNotFound)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					log.Printf("Error get gauge %s", err)
				}
				return
			} else {
				metricValue = fmt.Sprintf("%v", gauge)
			}

		case constants.MetricTypeCounter:
			if count, err := s.GetCounter(metricKey); err != nil {
				if errors.Is(err, myErr.ErrNotExist) {
					w.WriteHeader(http.StatusNotFound)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					log.Printf("Error get counter %s", err)
				}
				return
			} else {
				metricValue = fmt.Sprintf("%v", count)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			log.Printf("Error: unknown metric type '%s'", metricKey)
			return
		}

		// if ok
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(metricValue)); err != nil {
			log.Printf("Error: %s", err)
		}
	}
}

func GetListMetrics(s *service.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		html, err := s.GetListHTMLPage()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Error get html page %s", err)
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(html); err != nil {
			log.Printf("Error: %s", err)
		}
	}
}
