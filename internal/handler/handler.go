package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MrSwed/go-musthave-metrics/internal/constants"
	"github.com/MrSwed/go-musthave-metrics/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	s *service.Service
	r *chi.Mux
}

func NewHandler(s *service.Service) *Handler { return &Handler{s: s} }

func (h *Handler) InitRoutes() *Handler {
	h.r = chi.NewRouter()
	h.r.Use(middleware.Logger)

	h.r.Route("/", func(r chi.Router) {
		r.Get("/", h.GetListMetrics())
	})

	h.r.Route(fmt.Sprintf("%s/{%s}/{%s}/{%s}",
		constants.UpdateRoute, constants.MetricTypeParam, constants.MetricNameParam, constants.MetricValueParam),
		func(r chi.Router) { r.Post("/", h.UpdateMetric()) })

	h.r.Route(fmt.Sprintf("%s/{%s}/{%s}",
		constants.ValueRoute, constants.MetricTypeParam, constants.MetricNameParam),
		func(r chi.Router) {
			r.Get("/", h.GetMetric())
		})
	return h
}

func (h *Handler) RunServer(addr string) {
	if h.r == nil {
		h.InitRoutes()
	}
	log.Fatal(http.ListenAndServe(addr, h.r))
}
