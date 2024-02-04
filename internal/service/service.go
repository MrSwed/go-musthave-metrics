package service

import (
	"github.com/MrSwed/go-musthave-metrics/internal/config"
	"github.com/MrSwed/go-musthave-metrics/internal/repository"
)

type Service struct {
	Metrics
}

func NewService(r repository.Repository, c *config.StorageConfig) *Service {
	return &Service{Metrics: NewMetricService(r, c)}
}
