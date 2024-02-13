package repository

import (
	"errors"
	"sync"

	"github.com/MrSwed/go-musthave-metrics/internal/constant"
	"github.com/MrSwed/go-musthave-metrics/internal/domain"
	myErr "github.com/MrSwed/go-musthave-metrics/internal/errors"
)

type MemStorage interface {
	MemStore() *MemStorageRepo
}

type MemStorageCounter struct {
	Counter domain.Counters `json:"counter"`
	mc      sync.RWMutex
}

type MemStorageGauge struct {
	Gauge domain.Gauges `json:"gauge"`
	mg    sync.RWMutex
}

type MemStorageRepo struct {
	MemStorageCounter
	MemStorageGauge
}

func NewMemRepository() *MemStorageRepo {
	return &MemStorageRepo{
		MemStorageCounter: MemStorageCounter{Counter: domain.Counters{}},
		MemStorageGauge:   MemStorageGauge{Gauge: domain.Gauges{}},
	}
}

func (r *MemStorageRepo) MemStore() *MemStorageRepo {
	return r
}

func (r *MemStorageRepo) SetGauge(k string, v domain.Gauge) (err error) {
	r.mg.Lock()
	defer r.mg.Unlock()
	r.Gauge[k] = v
	return
}

func (r *MemStorageRepo) SetCounter(k string, v domain.Counter) (err error) {
	r.mc.Lock()
	defer r.mc.Unlock()
	r.Counter[k] = v
	return
}

func (r *MemStorageRepo) GetGauge(k string) (v domain.Gauge, err error) {
	var ok bool
	r.mg.RLock()
	defer r.mg.RUnlock()
	if v, ok = r.Gauge[k]; !ok {
		err = myErr.ErrNotExist
	}
	return
}

func (r *MemStorageRepo) GetCounter(k string) (v domain.Counter, err error) {
	var ok bool
	r.mc.RLock()
	defer r.mc.RUnlock()
	if v, ok = r.Counter[k]; !ok {
		err = myErr.ErrNotExist
	}
	return
}

func (r *MemStorageRepo) GetAllGauges() (domain.Gauges, error) {
	var err error
	return r.Gauge, err
}

func (r *MemStorageRepo) GetAllCounters() (domain.Counters, error) {
	var err error
	return r.Counter, err
}

func (r *MemStorageRepo) SetMetrics(metrics []domain.Metric) (newMetrics []domain.Metric, err error) {
	for _, metric := range metrics {
		switch metric.MType {
		case constant.MetricTypeGauge:
			if err = r.SetGauge(metric.ID, *metric.Value); err != nil {
				return
			}
			newMetrics = append(newMetrics, metric)
		case constant.MetricTypeCounter:
			var current domain.Counter
			if current, err = r.GetCounter(metric.ID); err != nil && !errors.Is(err, myErr.ErrNotExist) {
				return
			}
			delta := current + *metric.Delta
			if err = r.SetCounter(metric.ID, delta); err != nil {
				return
			}
			newMetrics = append(newMetrics, domain.Metric{
				ID:    metric.ID,
				MType: metric.MType,
				Delta: &delta,
			})
		}
	}
	return
}
