package service

import (
	"github.com/MrSwed/go-musthave-metrics/internal/constant"
	"github.com/MrSwed/go-musthave-metrics/internal/domain"
	"github.com/MrSwed/go-musthave-metrics/internal/helper"
	"github.com/MrSwed/go-musthave-metrics/internal/repository"
)

type MetricsHTML interface {
	GetCountersHTMLPage() ([]byte, error)
}

type MetricsHTMLService struct {
	r repository.Repository
}

func NewMetricsHTMLService(r repository.Repository) *MetricsHTMLService {
	return &MetricsHTMLService{r: r}
}

func (s *MetricsHTMLService) GetCountersHTMLPage() (html []byte, err error) {
	type lItem struct {
		MType  string
		MValue interface{}
	}
	var (
		counter domain.Counters
		gauge   domain.Gauges
		list    = map[string]lItem{}
	)
	if counter, err = s.r.GetAllCounters(); err != nil {
		return
	}
	if gauge, err = s.r.GetAllGauges(); err != nil {
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
	html, err = helper.ParseEmailHTMLTemplate(constant.MetricListTpl, list)
	return
}
