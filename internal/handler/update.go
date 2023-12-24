package handler

import (
	"github.com/MrSwed/go-musthave-metrics/internal/constants"
	"github.com/MrSwed/go-musthave-metrics/internal/service"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func UpdateMetric(s *service.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		params := strings.Split(
			strings.Trim(
				strings.TrimPrefix(r.URL.Path, constants.UpdateRoute),
				"/"),
			"/")
		if len(params) != 3 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		action, metricKey, metricValStr := params[0], params[1], params[2]
		switch action {
		case constants.MetricTypeGauge:
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
		case constants.MetricTypeCounter:
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
		default:
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("Error: unknown metric type '%s'", metricKey)
			return
		}

		// if ok
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("Saved: Ok")); err != nil {
			log.Printf("Error: %s", err)
		}
	}
}
