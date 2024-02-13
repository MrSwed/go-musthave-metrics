package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/MrSwed/go-musthave-metrics/internal/config"
	"github.com/MrSwed/go-musthave-metrics/internal/domain"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

func (h *Handler) UpdateMetric() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		action, metricKey, metricValStr := chi.URLParam(r, config.MetricTypeParam), chi.URLParam(r, config.MetricNameParam), chi.URLParam(r, config.MetricValueParam)
		switch action {
		case config.MetricTypeGauge:
			if v, err := domain.ParseGauge(metricValStr); err != nil {
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
			if v, err := domain.ParseCounter(metricValStr); err != nil {
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
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write([]byte("Bad input json")); err != nil {
				h.log.Error("Error return answer", zap.Error(err))
			}
			return
		}
		if metric, err = h.s.SetMetric(metric); err != nil {
			if errors.As(err, &validator.ValidationErrors{}) {
				w.WriteHeader(http.StatusBadRequest)
				if _, err := w.Write([]byte("Bad input data: " + err.Error())); err != nil {
					h.log.Error("Error return answer", zap.Error(err))
				}
			} else {
				h.log.Error("Error set metric", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
		var out []byte
		if out, err = json.Marshal(metric); err != nil {
			h.log.Error("Error marshal metric", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(out); err != nil {
			h.log.Error("Error return answer", zap.Error(err))
		}
	}
}

func (h *Handler) UpdateMetrics() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics []domain.Metric
		err := json.NewDecoder(r.Body).Decode(&metrics)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write([]byte("Bad input json")); err != nil {
				h.log.Error("Error return answer", zap.Error(err))
			}
			return
		}
		if metrics, err = h.s.SetMetrics(metrics); err != nil {
			if errors.As(err, &validator.ValidationErrors{}) {
				w.WriteHeader(http.StatusBadRequest)
				if _, err := w.Write([]byte("Bad input data: " + err.Error())); err != nil {
					h.log.Error("Error return answer", zap.Error(err))
				}
			} else {
				h.log.Error("Error set metric", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
		var out []byte
		if out, err = json.Marshal(metrics); err != nil {
			h.log.Error("Error marshal metrics", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(out); err != nil {
			h.log.Error("Error return answer", zap.Error(err))
		}
	}
}
