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
	SaveToFile() error
	RestoreFromFile() error
}

type MetricsService struct {
	r repository.Repository
	c *config.StorageConfig
}

func NewMetricService(r repository.Repository, c *config.StorageConfig) *MetricsService {
	return &MetricsService{r: r, c: c}
}

func (s *MetricsService) SetGauge(k string, v float64) (err error) {
	if err = s.r.SetGauge(k, v); err != nil {
		return
	}
	if s.c.StoreInterval == 0 {
		err = s.r.SaveToFile(s.r.MemStore())
	}
	return
}

func (s *MetricsService) IncreaseCounter(k string, v int64) (err error) {
	var prev int64
	if prev, err = s.r.GetCounter(k); err != nil && !errors.Is(err, myErr.ErrNotExist) {
		return
	} else {
		if err = s.r.SetCounter(k, prev+v); err != nil {
			return
		}
		if s.c.StoreInterval == 0 {
			err = s.r.SaveToFile(s.r.MemStore())
		}

		return
	}
}

func (s *MetricsService) GetGauge(k string) (v float64, err error) {
	v, err = s.r.GetGauge(k)
	return
}

func (s *MetricsService) GetCounter(k string) (v int64, err error) {
	v, err = s.r.GetCounter(k)
	return
}

func (s *MetricsService) GetCountersHTMLPage() (html []byte, err error) {
	type lItem struct {
		MType  string
		MValue interface{}
	}
	var (
		counter map[string]int64
		gauge   map[string]float64
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

func (s *MetricsService) SaveToFile() (err error) {
	return s.r.SaveToFile(s.r.MemStore())
}
func (s *MetricsService) RestoreFromFile() (err error) {
	return s.r.RestoreFromFile(s.r.MemStore())
}
