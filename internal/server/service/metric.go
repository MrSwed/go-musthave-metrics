package service

import (
	"errors"

	"golang.org/x/net/context"

	"go-musthave-metrics/internal/server/config"
	"go-musthave-metrics/internal/server/constant"
	"go-musthave-metrics/internal/server/domain"
	myErr "go-musthave-metrics/internal/server/errors"
	"go-musthave-metrics/internal/server/repository"

	"github.com/go-playground/validator/v10"
)

type Metrics interface {
	SetGauge(ctx context.Context, k string, v domain.Gauge) error
	IncreaseCounter(ctx context.Context, k string, v domain.Counter) error
	GetMetric(ctx context.Context, mType, k string) (domain.Metric, error)
	SetMetric(ctx context.Context, metric domain.Metric) (domain.Metric, error)
	SetMetrics(ctx context.Context, metrics []domain.Metric) ([]domain.Metric, error)
}

type MetricsService struct {
	r repository.Repository
	c *config.StorageConfig
}

func NewMetricService(r repository.Repository, c *config.StorageConfig) *MetricsService {
	return &MetricsService{r: r, c: c}
}

// SetGauge save one gauge
func (s *MetricsService) SetGauge(ctx context.Context, k string, v domain.Gauge) (err error) {
	if err = s.r.SetGauge(ctx, k, v); err != nil {
		return
	}
	if s.c.FileStoragePath != "" && s.c.FileStoreInterval == 0 {
		if _, err = s.SaveToFile(ctx); errors.Is(err, myErr.ErrNotMemMode) {
			err = nil
		}
	}
	return
}

// IncreaseCounter set or increase counter if already exist
func (s *MetricsService) IncreaseCounter(ctx context.Context, k string, v domain.Counter) (err error) {
	var prev domain.Counter
	if prev, err = s.r.GetCounter(ctx, k); err != nil && !errors.Is(err, myErr.ErrNotExist) {
		return
	} else {
		if err = s.r.SetCounter(ctx, k, prev+v); err != nil {
			return
		}
		if s.c.FileStoragePath != "" && s.c.FileStoreInterval == 0 {
			if _, err = s.SaveToFile(ctx); errors.Is(err, myErr.ErrNotMemMode) {
				err = nil
			}
		}

		return
	}
}

// GetMetric
// get counter or gauge depends type of metric
func (s *MetricsService) GetMetric(ctx context.Context, mType, id string) (v domain.Metric, err error) {
	switch mType {
	case constant.MetricTypeGauge:
		if gauge, er := s.r.GetGauge(ctx, id); er != nil {
			err = er
			return
		} else {
			v.Value = &gauge
		}
	case constant.MetricTypeCounter:
		if count, er := s.r.GetCounter(ctx, id); er != nil {
			err = er
			return
		} else {
			v.Delta = &count
		}
	default:
		err = myErr.ErrNotExist
		return
	}
	v.ID = id
	v.MType = mType
	return
}

// SetMetric set one metric
func (s *MetricsService) SetMetric(ctx context.Context, metric domain.Metric) (rm domain.Metric, err error) {
	validate := validator.New()
	if err = validate.Struct(metric); err != nil {
		return
	}
	switch metric.MType {
	case constant.MetricTypeGauge:
		if err = s.SetGauge(ctx, metric.ID, *metric.Value); err != nil {
			return
		}
	case constant.MetricTypeCounter:
		if err = s.IncreaseCounter(ctx, metric.ID, *metric.Delta); err != nil {
			return
		}
		var count domain.Counter
		if count, err = s.r.GetCounter(ctx, metric.ID); err != nil {
			return
		} else {
			metric.Delta = &count
		}
	}
	rm = metric
	if s.c.FileStoragePath != "" && s.c.FileStoreInterval == 0 {
		if _, err = s.SaveToFile(ctx); errors.Is(err, myErr.ErrNotMemMode) {
			err = nil
		}
	}

	return
}

// SetMetrics set several metrics
func (s *MetricsService) SetMetrics(ctx context.Context, metrics []domain.Metric) (rMetrics []domain.Metric, err error) {
	validate := validator.New()
	if err = validate.Struct(domain.ValidateMetrics{Metrics: metrics}); err != nil {
		return
	}
	rMetrics, err = s.r.SetMetrics(ctx, metrics)
	if s.c.FileStoragePath != "" && s.c.FileStoreInterval == 0 {
		if _, err = s.SaveToFile(ctx); errors.Is(err, myErr.ErrNotMemMode) {
			err = nil
		}
	}
	return
}
