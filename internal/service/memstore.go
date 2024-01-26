package service

import (
	"errors"
	"github.com/MrSwed/go-musthave-metrics/internal/config"
	myErr "github.com/MrSwed/go-musthave-metrics/internal/errors"
	"github.com/MrSwed/go-musthave-metrics/internal/helper"
	"github.com/MrSwed/go-musthave-metrics/internal/repository"
)

type Metrics interface {
	SetGauge(k string, v float64) error
	IncreaseCounter(k string, v int64) error
	GetGauge(k string) (float64, error)
	GetCounter(k string) (int64, error)
	GetCountersHTMLPage() ([]byte, error)
}

type MetricsService struct {
	r repository.MemStorage
}

func NewMemStorage(r repository.MemStorage) *MetricsService {
	return &MetricsService{r: r}
}

func (m *MetricsService) SetGauge(k string, v float64) error {
	return m.r.SetGauge(k, v)
}

func (m *MetricsService) IncreaseCounter(k string, v int64) error {
	if prev, err := m.r.GetCounter(k); err != nil && !errors.Is(err, myErr.ErrNotExist) {
		return err
	} else {
		return m.r.SetCounter(k, prev+v)
	}
}

func (m *MetricsService) GetGauge(k string) (v float64, err error) {
	v, err = m.r.GetGauge(k)
	return
}

func (m *MetricsService) GetCounter(k string) (v int64, err error) {
	v, err = m.r.GetCounter(k)
	return
}

func (m *MetricsService) GetCountersHTMLPage() (html []byte, err error) {
	type lItem struct {
		MType  string
		MValue interface{}
	}
	var (
		counter map[string]int64
		gauge   map[string]float64
		list    = map[string]lItem{}
	)
	if counter, err = m.r.GetAllCounters(); err != nil {
		return
	}
	if gauge, err = m.r.GetAllGauges(); err != nil {
		return
	}
	for k, v := range counter {
		list[k] = lItem{
			MType:  config.MetricTypeCounter,
			MValue: v,
		}
	}
	for k, v := range gauge {
		list[k] = lItem{
			MType:  config.MetricTypeGauge,
			MValue: v,
		}
	}
	html, err = helper.ParseEmailHTMLTemplate(config.MetricListTpl, list)
	return
}
