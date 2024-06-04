package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/MrSwed/go-musthave-metrics/internal/server/constant"
	"github.com/MrSwed/go-musthave-metrics/internal/server/domain"
	myErr "github.com/MrSwed/go-musthave-metrics/internal/server/errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// GetMetric
// get one metric value
//
//	GET http://server:port/value/metricType/metricName
func (h *Handler) GetMetric() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		mType, mKey := chi.URLParam(r, constant.MetricTypeParam), chi.URLParam(r, constant.MetricNameParam)
		ctx, cancel := context.WithTimeout(r.Context(), constant.ServerOperationTimeout*time.Second)
		defer cancel()
		metric, err := h.s.GetMetric(ctx, mType, mKey)
		if err != nil {
			if errors.Is(err, myErr.ErrNotExist) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				h.log.Error("Error get "+metric.MType, zap.Error(err))
			}
			return
		}

		setHeaderSHA(w, h.c.Key, []byte(metric.String()))
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(metric.String())); err != nil {
			h.log.Error("Error return answer", zap.Error(err))
		}
	}
}

// GetMetricJSON
// get one metric by json body
//
//	POST http://server:port/value
//	BODY {"id":metricName,"type":metricType}
func (h *Handler) GetMetricJSON() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric domain.Metric
		err := json.NewDecoder(r.Body).Decode(&metric)
		if err != nil || metric.ID == "" || metric.MType == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), constant.ServerOperationTimeout*time.Second)
		defer cancel()

		if metric, err = h.s.GetMetric(ctx, metric.MType, metric.ID); err != nil {
			if errors.Is(err, myErr.ErrNotExist) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				h.log.Error("Error get "+metric.MType, zap.Error(err))
			}
			return
		}

		var out []byte
		if out, err = json.Marshal(metric); err != nil {
			h.log.Error("Error marshal metric", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		setHeaderSHA(w, h.c.Key, out)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(out); err != nil {
			h.log.Error("Error return answer", zap.Error(err))
		}
	}
}

// GetListMetrics
// get html with all metrics
//
//	GET http://server:port/
func (h *Handler) GetListMetrics() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), constant.ServerOperationTimeout*time.Second)
		defer cancel()

		html, err := h.s.GetMetricsHTMLPage(ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.log.Error("Error get html page", zap.Error(err))
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		setHeaderSHA(w, h.c.Key, html)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(html); err != nil {
			h.log.Error("Error return answer", zap.Error(err))
		}
	}
}

// GetDBPing
// check is db ready
//
//	GET http://server:port/ping
func (h *Handler) GetDBPing() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), constant.ServerOperationTimeout*time.Second)
		defer cancel()
		if err := h.s.CheckDB(ctx); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.log.Error("Error ping", zap.Error(err))
			return
		}
		out := []byte("Status: ok")
		setHeaderSHA(w, h.c.Key, out)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(out); err != nil {
			h.log.Error("Error return answer", zap.Error(err))
		}
	}
}

// UpdateMetric
// update one metric from url params
//
//	POST http://server:port/update/metricType/metricName/metricValue
func (h *Handler) UpdateMetric() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		action, metricKey, metricValStr := chi.URLParam(r, constant.MetricTypeParam), chi.URLParam(r, constant.MetricNameParam), chi.URLParam(r, constant.MetricValueParam)
		ctx, cancel := context.WithTimeout(r.Context(), constant.ServerOperationTimeout*time.Second)
		defer cancel()

		switch action {
		case constant.MetricTypeGauge:
			if v, err := domain.ParseGauge(metricValStr); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				if _, er := w.Write([]byte("Bad metric value")); er != nil {
					h.log.Error("Error return answer", zap.Error(er))
				}
				return
			} else {
				if err = h.s.SetGauge(ctx, metricKey, v); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					h.log.Error("Error set gauge", zap.Error(err))
					return
				}
			}
		case constant.MetricTypeCounter:
			if v, er := domain.ParseCounter(metricValStr); er != nil {
				w.WriteHeader(http.StatusBadRequest)
				if _, err := w.Write([]byte("Bad metric value")); err != nil {
					h.log.Error("Error return answer", zap.Error(err))
				}
				return
			} else {
				if er = h.s.IncreaseCounter(ctx, metricKey, v); er != nil {
					w.WriteHeader(http.StatusInternalServerError)
					h.log.Error("Error set counter", zap.Error(er))
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
		out := []byte("Saved: Ok")
		setHeaderSHA(w, h.c.Key, out)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(out); err != nil {
			h.log.Error("Error return answer", zap.Error(err))
		}
	}
}

// UpdateMetricJSON
// update one metric from json body
//
//	POST http://server:port/update
//	BODY {"id":metricName,"type":metricType,"value":metricValue}
func (h *Handler) UpdateMetricJSON() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric domain.Metric
		err := json.NewDecoder(r.Body).Decode(&metric)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			if _, err = w.Write([]byte("Bad input json")); err != nil {
				h.log.Error("Error return answer", zap.Error(err))
			}
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), constant.ServerOperationTimeout*time.Second)
		defer cancel()

		if metric, err = h.s.SetMetric(ctx, metric); err != nil {
			if errors.As(err, &validator.ValidationErrors{}) {
				w.WriteHeader(http.StatusBadRequest)
				if _, err = w.Write([]byte("Bad input data: " + err.Error())); err != nil {
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
		setHeaderSHA(w, h.c.Key, out)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(out); err != nil {
			h.log.Error("Error return answer", zap.Error(err))
		}
	}
}

// UpdateMetrics
// update several metrics from json body
//
//	POST http://server:port/updates
//	BODY [{"id":metricName1,"type":metricType,"value":metricValue},{"id":metricName2,"type":metricType,"value":metricValue}]
func (h *Handler) UpdateMetrics() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics []domain.Metric
		err := json.NewDecoder(r.Body).Decode(&metrics)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			if _, err = w.Write([]byte("Bad input json")); err != nil {
				h.log.Error("Error return answer", zap.Error(err))
			}
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), constant.ServerOperationTimeout*time.Second)
		defer cancel()

		if metrics, err = h.s.SetMetrics(ctx, metrics); err != nil {
			if errors.As(err, &validator.ValidationErrors{}) {
				w.WriteHeader(http.StatusBadRequest)
				if _, err = w.Write([]byte("Bad input data: " + err.Error())); err != nil {
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
		setHeaderSHA(w, h.c.Key, out)
		w.WriteHeader(http.StatusOK)
		if _, er := w.Write(out); er != nil {
			h.log.Error("Error return answer", zap.Error(er))
		}
	}
}
