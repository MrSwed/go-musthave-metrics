package handler

import (
	"github.com/MrSwed/go-musthave-metrics/internal/constants"
	"github.com/MrSwed/go-musthave-metrics/internal/storage"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func UpdateMetric(w http.ResponseWriter, r *http.Request) {
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
			storage.Store.SetGauge(metricKey, v)
			log.Printf("Stored gauge %s = %f", metricKey, storage.Store.GetGauge(metricKey))
		}
	case constants.MetricTypeCounter:
		if v, err := strconv.ParseInt(metricValStr, 10, 64); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("Error: %s", err)
			return
		} else {
			storage.Store.IncreaseCounter(metricKey, v)
			log.Printf("Counter updated %s = %d", metricKey, storage.Store.GetCounter(metricKey))
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
