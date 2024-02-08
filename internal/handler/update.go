package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/MrSwed/go-musthave-metrics/internal/config"
	"github.com/MrSwed/go-musthave-metrics/internal/domain"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (h *Handler) UpdateMetric() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		action, metricKey, metricValStr := chi.URLParam(r, config.MetricTypeParam), chi.URLParam(r, config.MetricNameParam), chi.URLParam(r, config.MetricValueParam)
		switch action {
		case config.MetricTypeGauge:
			if v, err := strconv.ParseFloat(metricValStr, 64); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				if _, err := w.Write([]byte("Bad metric value")); err != nil {
					h.log.Error("Error return answer", zap.Error(err))
				}
				return
			} else {
				if err = h.s.SetGauge(metricKey, v); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					h.log.Error("Error set gauge", zap.Error(err))
					return
				}
			}
		case config.MetricTypeCounter:
			if v, err := strconv.ParseInt(metricValStr, 10, 64); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				if _, err := w.Write([]byte("Bad metric value")); err != nil {
					h.log.Error("Error return answer", zap.Error(err))
				}
				return
			} else {
				if err = h.s.IncreaseCounter(metricKey, v); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					h.log.Error("Error set counter", zap.Error(err))
					return
				}
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write([]byte("Unknown metric type")); err != nil {
				h.log.Error("Error return answer", zap.Error(err))
			}
			return
		}

		// if ok
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("Saved: Ok")); err != nil {
			h.log.Error("Error return answer", zap.Error(err))
		}
	}
}

func (h *Handler) UpdateMetricJSON() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric domain.Metric
		err := json.NewDecoder(r.Body).Decode(&metric)
		if err != nil || metric.ID == "" || metric.MType == "" {
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write([]byte("Bad metric id or type")); err != nil {
				h.log.Error("Error return answer", zap.Error(err))
			}
			return
		}
		switch metric.MType {
		case config.MetricTypeGauge:
			if metric.Value == nil {
				w.WriteHeader(http.StatusBadRequest)
				if _, err := w.Write([]byte("Bad metric value")); err != nil {
					h.log.Error("Error return answer", zap.Error(err))
				}
				return
			} else {
				if err = h.s.SetGauge(metric.ID, *metric.Value); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					h.log.Error("Error set gauge", zap.Error(err))
					return
				}
			}
		case config.MetricTypeCounter:
			if metric.Delta == nil {
				w.WriteHeader(http.StatusBadRequest)
				if _, err := w.Write([]byte("Bad metric value")); err != nil {
					h.log.Error("Error return answer", zap.Error(err))
				}
				return
			} else {
				if err = h.s.IncreaseCounter(metric.ID, *metric.Delta); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					h.log.Error("Error set counter", zap.Error(err))
					return
				}
				if count, err := h.s.GetCounter(metric.ID); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					h.log.Error("Error get counter", zap.Error(err))
					return
				} else {
					metric.Delta = &count
				}
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write([]byte("Unknown metric type")); err != nil {
				h.log.Error("Error return answer", zap.Error(err))
			}
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
