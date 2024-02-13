package service

import (
	"github.com/MrSwed/go-musthave-metrics/internal/config"
	"github.com/MrSwed/go-musthave-metrics/internal/repository"
)

type Service struct {
	Metrics
	MetricsHTML
	MetricsDB
	MetricsFile
}

func NewService(r repository.Repository, c *config.StorageConfig) *Service {
	mainService := NewMetricService(r, c)
	return &Service{
		Metrics:     mainService,
		MetricsHTML: NewMetricsHTMLService(r),
		MetricsDB:   NewMetricDBService(r),
		MetricsFile: mainService,
	}
}
