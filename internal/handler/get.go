package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/MrSwed/go-musthave-metrics/internal/constants"
	myErr "github.com/MrSwed/go-musthave-metrics/internal/errors"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetMetric() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		action, metricKey := chi.URLParam(r, constants.MetricTypeParam), chi.URLParam(r, constants.MetricNameParam)
		var metricValue string
		switch action {
		case constants.MetricTypeGauge:
			if gauge, err := h.s.GetGauge(metricKey); err != nil {
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
			if count, err := h.s.GetCounter(metricKey); err != nil {
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

func (h *Handler) GetListMetrics() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		html, err := h.s.GetCountersHTMLPage()
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
