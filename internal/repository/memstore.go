package repository

import (
	myErr "github.com/MrSwed/go-musthave-metrics/internal/errors"
	"sync"
)

type MemStorage interface {
	SetGauge(k string, v float64) error
	SetCounter(k string, v int64) error
	GetGauge(k string) (float64, error)
	GetCounter(k string) (int64, error)
	GetAllCounters() (map[string]int64, error)
	GetAllGauges() (map[string]float64, error)
	MemStore() *MemStorageRepository
}

type MemStorageCounter struct {
	Counter map[string]int64 `json:"counter"`
	mc      sync.RWMutex
}

type MemStorageGauge struct {
	Gauge map[string]float64 `json:"gauge"`
	mg    sync.RWMutex
}

type MemStorageRepository struct {
	MemStorageCounter
	MemStorageGauge
}

func NewMemRepository() *MemStorageRepository {
	return &MemStorageRepository{
		MemStorageCounter: MemStorageCounter{Counter: map[string]int64{}},
		MemStorageGauge:   MemStorageGauge{Gauge: map[string]float64{}},
	}
}

func (m *MemStorageRepository) MemStore() *MemStorageRepository {
	return m
}

func (m *MemStorageRepository) SetGauge(k string, v float64) (err error) {
	m.mg.Lock()
	defer m.mg.Unlock()
	m.Gauge[k] = v
	return
}

func (m *MemStorageRepository) SetCounter(k string, v int64) (err error) {
	m.mc.Lock()
	defer m.mc.Unlock()
	m.Counter[k] = v
	return
}

func (m *MemStorageRepository) GetGauge(k string) (v float64, err error) {
	var ok bool
	m.mg.RLock()
	defer m.mg.RUnlock()
	if v, ok = m.Gauge[k]; !ok {
		err = myErr.ErrNotExist
	}
	return
}

func (m *MemStorageRepository) GetCounter(k string) (v int64, err error) {
	var ok bool
	m.mc.RLock()
	defer m.mc.RUnlock()
	if v, ok = m.Counter[k]; !ok {
		err = myErr.ErrNotExist
	}
	return
}

func (m *MemStorageRepository) GetAllGauges() (map[string]float64, error) {
	var err error
	return m.Gauge, err
}

func (m *MemStorageRepository) GetAllCounters() (map[string]int64, error) {
	var err error
	return m.Counter, err
}
