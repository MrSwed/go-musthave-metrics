package service

import "github.com/MrSwed/go-musthave-metrics/internal/repository"

type Metrics interface {
	SetGauge(k string, v float64) error
	IncreaseCounter(k string, v int64) error
	GetGauge(k string) (float64, error)
	GetCounter(k string) (int64, error)
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
	if prev, err := m.r.GetCounter(k); err != nil {
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
