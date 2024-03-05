package handler

import (
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/MrSwed/go-musthave-metrics/internal/server/config"
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
	c   *config.WEB
}

func NewHandler(s *service.Service, c *config.WEB, log *zap.Logger) *Handler {
	return &Handler{s: s, c: c, log: log}
}

func signData(key string, data []byte) string {
	if key == "" {
		return ""
	}
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func setHeaderSHA(r http.ResponseWriter, key string, data []byte) {
	var sign string
	if sign = signData(key, data); sign == "" {
		return
	}
	r.Header().Set("HashSHA256", sign)
}

func (h *Handler) Handler() http.Handler {
	h.r = chi.NewRouter()
	h.r.Use(logger.Logger(h.log))
	h.r.Use(middleware.Compress(gzip.DefaultCompression, "application/json", "text/html"))
	h.r.Use(myMiddleware.Decompress(h.log))
	h.r.Use(myMiddleware.CheckSign(h.c, h.log))

	h.r.With(myMiddleware.TextHeader()).Route("/", func(r chi.Router) {
		r.Get("/", h.GetListMetrics())
		r.Get("/ping", h.GetDBPing())
	})

	h.r.Route(constant.UpdateRoute, func(r chi.Router) {
		r.With(myMiddleware.TextHeader()).Post(fmt.Sprintf("/{%s}/{%s}/{%s}",
			constant.MetricTypeParam, constant.MetricNameParam, constant.MetricValueParam),
			h.UpdateMetric())

		r.With(myMiddleware.JSONHeader()).Post("/", h.UpdateMetricJSON())
	})

	h.r.Route(constant.UpdatesRoute, func(r chi.Router) {
		r.With(myMiddleware.JSONHeader()).Post("/", h.UpdateMetrics())
	})

	h.r.Route(constant.ValueRoute, func(r chi.Router) {
		r.With(myMiddleware.TextHeader()).Get(fmt.Sprintf("/{%s}/{%s}",
			constant.MetricTypeParam, constant.MetricNameParam), h.GetMetric())

		r.With(myMiddleware.JSONHeader()).Post("/", h.GetMetricJSON())
	})

	return h.r
}
