package handler

import (
	"fmt"
	"net/http"

	"github.com/MrSwed/go-musthave-metrics/internal/constants"
	"github.com/MrSwed/go-musthave-metrics/internal/logger"
	"github.com/MrSwed/go-musthave-metrics/internal/service"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Handler struct {
	s   *service.Service
	r   *chi.Mux
	log *zap.Logger
}

func NewHandler(s *service.Service, log *zap.Logger) *Handler { return &Handler{s: s, log: log} }

func (h *Handler) InitRoutes() *Handler {
	h.r = chi.NewRouter()
	h.r.Use(logger.Logger(h.log))

	h.r.Route("/", func(r chi.Router) {
		r.Get("/", h.GetListMetrics())
	})

	h.r.Route(constants.UpdateRoute, func(r chi.Router) {
		r.Post(fmt.Sprintf("/{%s}/{%s}/{%s}",
			constants.MetricTypeParam, constants.MetricNameParam, constants.MetricValueParam),
			h.UpdateMetric())
		r.Post("/", h.UpdateMetricJSON())
	})

	h.r.Route(constants.ValueRoute, func(r chi.Router) {
		r.Get(fmt.Sprintf("/{%s}/{%s}",
			constants.MetricTypeParam, constants.MetricNameParam), h.GetMetric())
		r.Post("/", h.GetMetricJSON())
	})
	return h
}

func (h *Handler) RunServer(addr string) error {
	if h.r == nil {
		h.InitRoutes()
	}
	return http.ListenAndServe(addr, h.r)
}
