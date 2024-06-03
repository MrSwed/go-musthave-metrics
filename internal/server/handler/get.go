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
