package handler

import (
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"go-musthave-metrics/internal/server/config"
	"go-musthave-metrics/internal/server/constant"
	"go-musthave-metrics/internal/server/service"
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
	c   *config.Config
}

// NewHandler return app handler
func NewHandler(s *service.Service, c *config.Config, log *zap.Logger) *Handler {
	return &Handler{
		app: chi.NewRouter(),
		s:   s,
		c:   c,
		log: log}
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
func (h *Handler) HTTPHandler() http.Handler {
	h.app.Use(Logger(h.log))
	h.app.Use(middleware.Compress(gzip.DefaultCompression, "application/json", "text/html"))
	h.app.Use(Decrypt(h.c.GetPrivateKey(), h.log))
	h.app.Use(Decompress(h.log))
	h.app.Use(CheckSign(&h.c.WEB, h.log))
	h.app.Use(CheckNetwork(&h.c.WEB, h.log))

	h.app.Mount("/debug", middleware.Profiler())

	h.app.With(TextHeader()).Route("/", func(r chi.Router) {
		r.Get("/", h.GetListMetrics())
		r.Get("/ping", h.GetDBPing())
	})

	h.app.Route(constant.UpdateRoute, func(r chi.Router) {
		r.With(TextHeader()).Post(fmt.Sprintf("/{%s}/{%s}/{%s}",
			constant.MetricTypeParam, constant.MetricNameParam, constant.MetricValueParam),
			h.UpdateMetric())

		r.With(JSONHeader()).Post("/", h.UpdateMetricJSON())
	})

	h.app.Route(constant.UpdatesRoute, func(r chi.Router) {
		r.With(JSONHeader()).Post("/", h.UpdateMetrics())
	})

	h.app.Route(constant.ValueRoute, func(r chi.Router) {
		r.With(TextHeader()).Get(fmt.Sprintf("/{%s}/{%s}",
			constant.MetricTypeParam, constant.MetricNameParam), h.GetMetric())

		r.With(JSONHeader()).Post("/", h.GetMetricJSON())
	})

	return h.app
}
