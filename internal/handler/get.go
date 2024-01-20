package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/MrSwed/go-musthave-metrics/internal/constants"
	"github.com/MrSwed/go-musthave-metrics/internal/domain"
	myErr "github.com/MrSwed/go-musthave-metrics/internal/errors"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
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
					h.log.Error("Error get gauge", zap.Error(err))
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
					h.log.Error("Error get counter", zap.Error(err))
				}
				return
			} else {
				metricValue = fmt.Sprintf("%v", count)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// if ok
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(metricValue)); err != nil {
			h.log.Error("Error return answer", zap.Error(err))
		}
	}
}

func (h *Handler) GetMetricJSON() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric domain.Metric
		err := json.NewDecoder(r.Body).Decode(&metric)
		if err != nil || metric.ID == "" || metric.MType == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		switch metric.MType {
		case constants.MetricTypeGauge:
			if gauge, err := h.s.GetGauge(metric.ID); err != nil {
				if errors.Is(err, myErr.ErrNotExist) {
					w.WriteHeader(http.StatusNotFound)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					h.log.Error("Error get gauge", zap.Error(err))
				}
				return
			} else {
				metric.Value = &gauge
			}
		case constants.MetricTypeCounter:
			if count, err := h.s.GetCounter(metric.ID); err != nil {
				if errors.Is(err, myErr.ErrNotExist) {
					w.WriteHeader(http.StatusNotFound)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					h.log.Error("Error get counter", zap.Error(err))
				}
				return
			} else {
				metric.Delta = &count
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var out []byte
		if out, err = json.Marshal(metric); err != nil {
			h.log.Error("Error marshal metric", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(out); err != nil {
			h.log.Error("Error return answer", zap.Error(err))
		}
	}
}

func (h *Handler) GetListMetrics() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		html, err := h.s.GetCountersHTMLPage()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.log.Error("Error get html page", zap.Error(err))
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(html); err != nil {
			h.log.Error("Error return answer", zap.Error(err))
		}
	}
}
