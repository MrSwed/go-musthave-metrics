package service

import (
	"github.com/MrSwed/go-musthave-metrics/internal/repository"
)

type MetricsDB interface {
	CheckDB() error
}

type MetricsDBService struct {
	r repository.Repository
}

func NewMetricDBService(r repository.Repository) *MetricsDBService {
	return &MetricsDBService{r: r}
}

func (s *MetricsDBService) CheckDB() error {
	return s.r.Ping()
}
