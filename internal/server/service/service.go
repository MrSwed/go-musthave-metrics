package service

import (
	"go-musthave-metrics/internal/server/config"
	"go-musthave-metrics/internal/server/repository"
)

type Service struct {
	Metrics
	MetricsHTML
	MetricsDB
	MetricsFile
}

// NewService return main service methods
func NewService(r repository.Repository, c *config.StorageConfig) *Service {
	mainService := NewMetricService(r, c)
	return &Service{
		Metrics:     mainService,
		MetricsHTML: NewMetricsHTMLService(r),
		MetricsDB:   NewMetricDBService(r),
		MetricsFile: mainService,
	}
}
