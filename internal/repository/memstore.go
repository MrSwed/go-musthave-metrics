package repository

import (
	"github.com/MrSwed/go-musthave-metrics/internal/errors"
)

type MemStorage interface {
	SetGauge(k string, v float64) error
	SetCounter(k string, v int64) error
	GetGauge(k string) (float64, error)
	GetCounter(k string) (int64, error)
	GetCountersList() (map[string]int64, error)
	GetGaugesList() (map[string]float64, error)
}

type MemStorageRepository struct {
	gauge   map[string]float64
	counter map[string]int64
}

func NewMemRepository() *MemStorageRepository {
	return &MemStorageRepository{gauge: map[string]float64{}, counter: map[string]int64{}}
}

func (m *MemStorageRepository) SetGauge(k string, v float64) (err error) {
	m.gauge[k] = v
	return
}

func (m *MemStorageRepository) SetCounter(k string, v int64) (err error) {
	m.counter[k] = v
	return
}

func (m *MemStorageRepository) GetGauge(k string) (v float64, err error) {
	var ok bool
	if v, ok = m.gauge[k]; !ok {
		err = errors.ErrNotExist
	}
	return
}

func (m *MemStorageRepository) GetCounter(k string) (v int64, err error) {
	var ok bool
	if v, ok = m.counter[k]; !ok {
		err = errors.ErrNotExist
	}
	return
}

func (m *MemStorageRepository) GetCountersList() (list map[string]int64, err error) {
	list = m.counter
	return
}
func (m *MemStorageRepository) GetGaugesList() (list map[string]float64, err error) {
	list = m.gauge
	return
}
