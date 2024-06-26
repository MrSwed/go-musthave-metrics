package service

import (
	"context"

	"go-musthave-metrics/internal/server/constant"
	"go-musthave-metrics/internal/server/domain"
	"go-musthave-metrics/internal/server/helper"
	"go-musthave-metrics/internal/server/repository"
)

type MetricsHTML interface {
	GetMetricsHTMLPage(ctx context.Context) ([]byte, error)
}

type MetricsHTMLService struct {
	r repository.Repository
}

func NewMetricsHTMLService(r repository.Repository) *MetricsHTMLService {
	return &MetricsHTMLService{r: r}
}

// GetMetricsHTMLPage get html page with all metrics
func (s *MetricsHTMLService) GetMetricsHTMLPage(ctx context.Context) (html []byte, err error) {
	type lItem struct {
		MValue interface{}
		MType  string
	}
	var (
		counter domain.Counters
		gauge   domain.Gauges
		list    = map[string]lItem{}
	)
	if counter, err = s.r.GetAllCounters(ctx); err != nil {
		return
	}
	if gauge, err = s.r.GetAllGauges(ctx); err != nil {
		return
	}
	for k, v := range counter {
		list[k] = lItem{
			MType:  constant.MetricTypeCounter,
			MValue: v,
		}
	}
	for k, v := range gauge {
		list[k] = lItem{
			MType:  constant.MetricTypeGauge,
			MValue: v,
		}
	}
	html, err = helper.ParseHTMLTemplate(constant.MetricListTpl, list)
	return
}
