package handler

import (
	"github.com/MrSwed/go-musthave-metrics/internal/service"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
)

func UpdateGaugeMetric(s *service.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		metricKey, metricValStr := chi.URLParam(r, "metricName"), chi.URLParam(r, "metricValue")
		if v, err := strconv.ParseFloat(metricValStr, 64); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("Error: %s", err)
			return
		} else {
			if err = s.SetGauge(metricKey, v); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Printf("Error set gauge %s", err)
			} else {
				newV, _ := s.GetGauge(metricKey)
				log.Printf("Stored gauge %s = %f", metricKey, newV)
			}
		}

		// if ok
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("Saved: Ok")); err != nil {
			log.Printf("Error: %s", err)
		}
	}
}

func UpdateCounterMetric(s *service.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricKey, metricValStr := chi.URLParam(r, "metricName"), chi.URLParam(r, "metricValue")
		if v, err := strconv.ParseInt(metricValStr, 10, 64); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("Error: %s", err)
			return
		} else {
			if err = s.IncreaseCounter(metricKey, v); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Printf("Error set counter %s", err)
			} else {
				newV, _ := s.GetCounter(metricKey)
				log.Printf("Counter updated %s = %d", metricKey, newV)
			}
		}
		// if ok
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("Saved: Ok")); err != nil {
			log.Printf("Error: %s", err)
		}
	}
}
