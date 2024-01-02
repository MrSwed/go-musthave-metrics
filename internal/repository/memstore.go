package repository

import (
	"github.com/MrSwed/go-musthave-metrics/internal/errors"
	"sync"
)

type MemStorage interface {
	SetGauge(k string, v float64) error
	SetCounter(k string, v int64) error
	GetGauge(k string) (float64, error)
	GetCounter(k string) (int64, error)
	GetAllCounters() (map[string]int64, error)
	GetAllGauges() (map[string]float64, error)
}

type MemStorageCounter struct {
	counter map[string]int64
	mc      sync.RWMutex
}

type MemStorageGauge struct {
	gauge map[string]float64
	mg    sync.RWMutex
}

type MemStorageRepository struct {
	MemStorageCounter
	MemStorageGauge
}

func NewMemRepository() *MemStorageRepository {
	return &MemStorageRepository{
		MemStorageCounter: MemStorageCounter{counter: map[string]int64{}},
		MemStorageGauge:   MemStorageGauge{gauge: map[string]float64{}},
	}
}

func (m *MemStorageRepository) SetGauge(k string, v float64) (err error) {
	m.mg.Lock()
	defer m.mg.Unlock()
	m.gauge[k] = v
	return
}

func (m *MemStorageRepository) SetCounter(k string, v int64) (err error) {
	m.mc.Lock()
	defer m.mc.Unlock()
	m.counter[k] = v
	return
}

func (m *MemStorageRepository) GetGauge(k string) (v float64, err error) {
	var ok bool
	m.mg.RLock()
	defer m.mg.RUnlock()
	if v, ok = m.gauge[k]; !ok {
		err = errors.ErrNotExist
	}
	return
}

func (m *MemStorageRepository) GetCounter(k string) (v int64, err error) {
	var ok bool
	m.mc.RLock()
	defer m.mc.RUnlock()
	if v, ok = m.counter[k]; !ok {
		err = errors.ErrNotExist
	}
	return
}

func (m *MemStorageRepository) GetAllGauges() (map[string]float64, error) {
	var err error
	return m.gauge, err
}

func (m *MemStorageRepository) GetAllCounters() (map[string]int64, error) {
	var err error
	return m.counter, err
}
