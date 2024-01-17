package handler

import (
	"net/http"
	"strconv"

	"github.com/MrSwed/go-musthave-metrics/internal/constants"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (h *Handler) UpdateMetric() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		action, metricKey, metricValStr := chi.URLParam(r, constants.MetricTypeParam), chi.URLParam(r, constants.MetricNameParam), chi.URLParam(r, constants.MetricValueParam)
		switch action {
		case constants.MetricTypeGauge:
			if v, err := strconv.ParseFloat(metricValStr, 64); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			} else {
				if err = h.s.SetGauge(metricKey, v); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					h.log.Error("Error set gauge", zap.Error(err))
				}
			}
		case constants.MetricTypeCounter:
			if v, err := strconv.ParseInt(metricValStr, 10, 64); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			} else {
				if err = h.s.IncreaseCounter(metricKey, v); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					h.log.Error("Error set counter", zap.Error(err))
				}
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
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
