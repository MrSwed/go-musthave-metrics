package handler

import (
	"compress/gzip"
	"fmt"
	"net/http"

	"github.com/MrSwed/go-musthave-metrics/internal/config"
	"github.com/MrSwed/go-musthave-metrics/internal/logger"
	myMiddleware "github.com/MrSwed/go-musthave-metrics/internal/middleware"
	"github.com/MrSwed/go-musthave-metrics/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type Handler struct {
	s   *service.Service
	r   *chi.Mux
	c   *config.ServerConfig
	log *zap.Logger
}

func NewHandler(s *service.Service, c *config.ServerConfig, log *zap.Logger) *Handler {
	return &Handler{s: s, c: c, log: log}
}

func (h *Handler) InitRoutes() *Handler {
	h.r = chi.NewRouter()
	h.r.Use(logger.Logger(h.log))
	h.r.Use(middleware.Compress(gzip.DefaultCompression, "application/json", "text/html"))
	h.r.Use(myMiddleware.Decompress)

	h.r.Route("/", func(r chi.Router) {
		r.Get("/", h.GetListMetrics())
	})

	h.r.Route(config.UpdateRoute, func(r chi.Router) {
		r.Post(fmt.Sprintf("/{%s}/{%s}/{%s}",
			config.MetricTypeParam, config.MetricNameParam, config.MetricValueParam),
			h.UpdateMetric())
		r.Post("/", h.UpdateMetricJSON())
	})

	h.r.Route(config.ValueRoute, func(r chi.Router) {
		r.Get(fmt.Sprintf("/{%s}/{%s}",
			config.MetricTypeParam, config.MetricNameParam), h.GetMetric())
		r.Post("/", h.GetMetricJSON())
	})
	return h
}

func (h *Handler) RunServer() error {
	if h.r == nil {
		h.InitRoutes()
	}
	return http.ListenAndServe(h.c.ServerAddress, h.r)
}
