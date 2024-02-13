package service

import (
	"errors"

	"github.com/MrSwed/go-musthave-metrics/internal/config"
	"github.com/MrSwed/go-musthave-metrics/internal/constant"
	"github.com/MrSwed/go-musthave-metrics/internal/domain"
	myErr "github.com/MrSwed/go-musthave-metrics/internal/errors"
	"github.com/MrSwed/go-musthave-metrics/internal/repository"

	"github.com/go-playground/validator/v10"
)

type Metrics interface {
	SetGauge(k string, v domain.Gauge) error
	IncreaseCounter(k string, v domain.Counter) error
	GetGauge(k string) (domain.Gauge, error)
	GetCounter(k string) (domain.Counter, error)
	SetMetric(metric domain.Metric) (domain.Metric, error)
	SetMetrics(metrics []domain.Metric) ([]domain.Metric, error)
}

type MetricsService struct {
	r repository.Repository
	c *config.StorageConfig
}

func NewMetricService(r repository.Repository, c *config.StorageConfig) *MetricsService {
	return &MetricsService{r: r, c: c}
}

func (s *MetricsService) SetGauge(k string, v domain.Gauge) (err error) {
	if err = s.r.SetGauge(k, v); err != nil {
		return
	}
	if s.c.FileStoragePath != "" && s.c.StoreInterval == 0 {
		if _, err = s.SaveToFile(); errors.Is(err, myErr.ErrNotMemMode) {
			err = nil
		}
	}
	return
}

func (s *MetricsService) IncreaseCounter(k string, v domain.Counter) (err error) {
	var prev domain.Counter
	if prev, err = s.r.GetCounter(k); err != nil && !errors.Is(err, myErr.ErrNotExist) {
		return
	} else {
		if err = s.r.SetCounter(k, prev+v); err != nil {
			return
		}
		if s.c.FileStoragePath != "" && s.c.StoreInterval == 0 {
			if _, err = s.SaveToFile(); errors.Is(err, myErr.ErrNotMemMode) {
				err = nil
			}
		}

		return
	}
}

func (s *MetricsService) GetGauge(k string) (v domain.Gauge, err error) {
	v, err = s.r.GetGauge(k)
	return
}

func (s *MetricsService) GetCounter(k string) (v domain.Counter, err error) {
	v, err = s.r.GetCounter(k)
	return
}

func (s *MetricsService) SetMetric(metric domain.Metric) (rm domain.Metric, err error) {
	validate := validator.New()
	if err = validate.Struct(metric); err != nil {
		return
	}
	switch metric.MType {
	case constant.MetricTypeGauge:
		if err = s.SetGauge(metric.ID, *metric.Value); err != nil {
			return
		}
	case constant.MetricTypeCounter:
		if err = s.IncreaseCounter(metric.ID, *metric.Delta); err != nil {
			return
		}
		var count domain.Counter
		if count, err = s.GetCounter(metric.ID); err != nil {
			return
		} else {
			metric.Delta = &count
		}
	}
	rm = metric
	if s.c.FileStoragePath != "" && s.c.StoreInterval == 0 {
		if _, err = s.SaveToFile(); errors.Is(err, myErr.ErrNotMemMode) {
			err = nil
		}
	}

	return
}

func (s *MetricsService) SetMetrics(metrics []domain.Metric) (rMetrics []domain.Metric, err error) {
	validate := validator.New()
	if err = validate.Struct(domain.ValidateMetrics{Metrics: metrics}); err != nil {
		return
	}
	rMetrics, err = s.r.SetMetrics(metrics)
	if s.c.FileStoragePath != "" && s.c.StoreInterval == 0 {
		if _, err = s.SaveToFile(); errors.Is(err, myErr.ErrNotMemMode) {
			err = nil
		}
	}
	return
}
