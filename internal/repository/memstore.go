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

func (r *MemStorageRepo) MemStore() *MemStorageRepo {
	return r
}

func (r *MemStorageRepo) SetGauge(k string, v float64) (err error) {
	r.mg.Lock()
	defer r.mg.Unlock()
	r.Gauge[k] = v
	return
}

func (r *MemStorageRepo) SetCounter(k string, v int64) (err error) {
	r.mc.Lock()
	defer r.mc.Unlock()
	r.Counter[k] = v
	return
}

func (r *MemStorageRepo) GetGauge(k string) (v float64, err error) {
	var ok bool
	r.mg.RLock()
	defer r.mg.RUnlock()
	if v, ok = r.Gauge[k]; !ok {
		err = myErr.ErrNotExist
	}
	return
}

func (r *MemStorageRepo) GetCounter(k string) (v int64, err error) {
	var ok bool
	r.mc.RLock()
	defer r.mc.RUnlock()
	if v, ok = r.Counter[k]; !ok {
		err = myErr.ErrNotExist
	}
	return
}

func (r *MemStorageRepo) GetAllGauges() (map[string]float64, error) {
	var err error
	return r.Gauge, err
}

func (r *MemStorageRepo) GetAllCounters() (map[string]int64, error) {
	var err error
	return r.Counter, err
}
