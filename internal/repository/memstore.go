package repository

import (
	"sync"

	myErr "github.com/MrSwed/go-musthave-metrics/internal/errors"
)

type MemStorage interface {
	MemStore() *MemStorageRepo
}

type MemStorageCounter struct {
	Counter map[string]int64 `json:"counter"`
	mc      sync.RWMutex
}

type MemStorageGauge struct {
	Gauge map[string]float64 `json:"gauge"`
	mg    sync.RWMutex
}

type MemStorageRepo struct {
	MemStorageCounter
	MemStorageGauge
}

func NewMemRepository() *MemStorageRepo {
	return &MemStorageRepo{
		MemStorageCounter: MemStorageCounter{Counter: map[string]int64{}},
		MemStorageGauge:   MemStorageGauge{Gauge: map[string]float64{}},
	}
}

func (m *MemStorageRepo) MemStore() *MemStorageRepo {
	return m
}

func (m *MemStorageRepo) SetGauge(k string, v float64) (err error) {
	m.mg.Lock()
	defer m.mg.Unlock()
	m.Gauge[k] = v
	return
}

func (m *MemStorageRepo) SetCounter(k string, v int64) (err error) {
	m.mc.Lock()
	defer m.mc.Unlock()
	m.Counter[k] = v
	return
}

func (m *MemStorageRepo) GetGauge(k string) (v float64, err error) {
	var ok bool
	m.mg.RLock()
	defer m.mg.RUnlock()
	if v, ok = m.Gauge[k]; !ok {
		err = myErr.ErrNotExist
	}
	return
}

func (m *MemStorageRepo) GetCounter(k string) (v int64, err error) {
	var ok bool
	m.mc.RLock()
	defer m.mc.RUnlock()
	if v, ok = m.Counter[k]; !ok {
		err = myErr.ErrNotExist
	}
	return
}

func (m *MemStorageRepo) GetAllGauges() (map[string]float64, error) {
	var err error
	return m.Gauge, err
}

func (m *MemStorageRepo) GetAllCounters() (map[string]int64, error) {
	var err error
	return m.Counter, err
}
