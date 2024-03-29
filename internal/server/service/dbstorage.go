package service

import (
	"context"

	"github.com/MrSwed/go-musthave-metrics/internal/server/repository"
)

type MetricsDB interface {
	CheckDB(ctx context.Context) error
}

type MetricsDBService struct {
	r repository.Repository
}

func NewMetricDBService(r repository.Repository) *MetricsDBService {
	return &MetricsDBService{r: r}
}

func (s *MetricsDBService) CheckDB(ctx context.Context) error {
	return s.r.Ping(ctx)
}
