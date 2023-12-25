package service

import (
	"errors"
	"github.com/MrSwed/go-musthave-metrics/internal/constants"
	myErr "github.com/MrSwed/go-musthave-metrics/internal/errors"
	"github.com/MrSwed/go-musthave-metrics/internal/helper"
	"github.com/MrSwed/go-musthave-metrics/internal/repository"
)

type Metrics interface {
	SetGauge(k string, v float64) error
	IncreaseCounter(k string, v int64) error
	GetGauge(k string) (float64, error)
	GetCounter(k string) (int64, error)
	GetListHtmlPage() ([]byte, error)
}

type MetricsService struct {
	r repository.MemStorage
}

func NewMemStorage(r repository.Repository) *MetricsService {
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

func (m *MetricsService) GetListHtmlPage() (html []byte, err error) {
	type lItem struct {
		MType  string
		MValue interface{}
	}
	var (
		counter map[string]int64
		gauge   map[string]float64
		list    = map[string]lItem{}
	)
	if counter, err = m.r.GetCountersList(); err != nil {
		return
	}
	if gauge, err = m.r.GetGaugesList(); err != nil {
		return
	}
	for k, v := range counter {
		list[k] = lItem{
			MType:  constants.MetricTypeCounter,
			MValue: v,
		}
	}
	for k, v := range gauge {
		list[k] = lItem{
			MType:  constants.MetricTypeGauge,
			MValue: v,
		}
	}
	html, err = helper.ParseEmailHtmlTemplate(constants.MetricListTpl, list)
	return
}
