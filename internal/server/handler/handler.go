package handler

import (
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/MrSwed/go-musthave-metrics/internal/server/config"
	"github.com/MrSwed/go-musthave-metrics/internal/server/constant"
	myMiddleware "github.com/MrSwed/go-musthave-metrics/internal/server/middleware"
	"github.com/MrSwed/go-musthave-metrics/internal/server/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// Handler
// main handler
type Handler struct {
	s   *service.Service
	app *chi.Mux
	log *zap.Logger
	c   *config.WEB
}

// NewHandler return app handler
func NewHandler(app *chi.Mux, s *service.Service, c *config.WEB, log *zap.Logger) *Handler {
	return &Handler{app: app, s: s, c: c, log: log}
}

func SignData(key string, data []byte) string {
	if key == "" {
		return ""
	}
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func setHeaderSHA(r http.ResponseWriter, key string, data []byte) {
	var sign string
	if sign = SignData(key, data); sign == "" {
		return
	}
	r.Header().Set(constant.HeaderSignKey, sign)
}

// Handler
// init app routes
func (h *Handler) Handler() http.Handler {
	h.app.Use(myMiddleware.Logger(h.log))
	h.app.Use(middleware.Compress(gzip.DefaultCompression, "application/json", "text/html"))
	h.app.Use(myMiddleware.Decrypt(h.c.GetPrivateKey(), h.log))
	h.app.Use(myMiddleware.Decompress(h.log))
	h.app.Use(myMiddleware.CheckSign(h.c, h.log))
	h.app.Use(myMiddleware.CheckNetwork(h.c, h.log))

	h.app.Mount("/debug", middleware.Profiler())

	h.app.With(myMiddleware.TextHeader()).Route("/", func(r chi.Router) {
		r.Get("/", h.GetListMetrics())
		r.Get("/ping", h.GetDBPing())
	})

	h.app.Route(constant.UpdateRoute, func(r chi.Router) {
		r.With(myMiddleware.TextHeader()).Post(fmt.Sprintf("/{%s}/{%s}/{%s}",
			constant.MetricTypeParam, constant.MetricNameParam, constant.MetricValueParam),
			h.UpdateMetric())

		r.With(myMiddleware.JSONHeader()).Post("/", h.UpdateMetricJSON())
	})

	h.app.Route(constant.UpdatesRoute, func(r chi.Router) {
		r.With(myMiddleware.JSONHeader()).Post("/", h.UpdateMetrics())
	})

	h.app.Route(constant.ValueRoute, func(r chi.Router) {
		r.With(myMiddleware.TextHeader()).Get(fmt.Sprintf("/{%s}/{%s}",
			constant.MetricTypeParam, constant.MetricNameParam), h.GetMetric())

		r.With(myMiddleware.JSONHeader()).Post("/", h.GetMetricJSON())
	})

	return h.app
}
