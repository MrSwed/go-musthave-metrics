package handler

import (
	"compress/gzip"
	"fmt"
	"net/http"

	"github.com/MrSwed/go-musthave-metrics/internal/server/constant"
	"github.com/MrSwed/go-musthave-metrics/internal/server/logger"
	myMiddleware "github.com/MrSwed/go-musthave-metrics/internal/server/middleware"
	"github.com/MrSwed/go-musthave-metrics/internal/server/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type Handler struct {
	s   *service.Service
	r   *chi.Mux
	log *zap.Logger
}

func NewHandler(s *service.Service, log *zap.Logger) *Handler {
	return &Handler{s: s, log: log}
}

func (h *Handler) Handler() http.Handler {
	h.r = chi.NewRouter()
	h.r.Use(logger.Logger(h.log))
	h.r.Use(middleware.Compress(gzip.DefaultCompression, "application/json", "text/html"))
	h.r.Use(myMiddleware.Decompress(h.log))

	h.r.Route("/", func(r chi.Router) {
		r.Use(myMiddleware.TextHeader())
		r.Get("/", h.GetListMetrics())
		r.Get("/ping", h.GetDBPing())
	})

	h.r.Route(constant.UpdateRoute, func(r chi.Router) {
		r.Use(myMiddleware.TextHeader())
		r.Post(fmt.Sprintf("/{%s}/{%s}/{%s}",
			constant.MetricTypeParam, constant.MetricNameParam, constant.MetricValueParam),
			h.UpdateMetric())

		r.Route("/", func(r chi.Router) {
			r.Use(myMiddleware.JSONHeader())
			r.Post("/", h.UpdateMetricJSON())
		})
	})
	h.r.Route(constant.UpdatesRoute, func(r chi.Router) {
		r.Use(myMiddleware.JSONHeader())
		r.Post("/", h.UpdateMetrics())
	})

	h.r.Route(constant.ValueRoute, func(r chi.Router) {
		r.Use(myMiddleware.TextHeader())
		r.Get(fmt.Sprintf("/{%s}/{%s}",
			constant.MetricTypeParam, constant.MetricNameParam), h.GetMetric())

		r.Route("/", func(r chi.Router) {
			r.Use(myMiddleware.JSONHeader())
			r.Post("/", h.GetMetricJSON())
		})
	})

	return h.r
}
