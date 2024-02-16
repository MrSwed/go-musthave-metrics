package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/MrSwed/go-musthave-metrics/internal/server/constant"
	"github.com/MrSwed/go-musthave-metrics/internal/server/domain"
	myErr "github.com/MrSwed/go-musthave-metrics/internal/server/errors"
)

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

func (r *MemStorageRepo) Ping() (err error) {
	return
}

func (r *MemStorageRepo) MemStore(ctx context.Context) (*MemStorageRepo, error) {
	return r, nil
}

func (r *MemStorageRepo) SetGauge(ctx context.Context, k string, v domain.Gauge) (err error) {
	r.mg.Lock()
	defer r.mg.Unlock()
	r.Gauge[k] = v
	return
}

func (r *MemStorageRepo) SetCounter(ctx context.Context, k string, v domain.Counter) (err error) {
	r.mc.Lock()
	defer r.mc.Unlock()
	r.Counter[k] = v
	return
}

func (r *MemStorageRepo) GetGauge(ctx context.Context, k string) (v domain.Gauge, err error) {
	var ok bool
	r.mg.RLock()
	defer r.mg.RUnlock()
	if v, ok = r.Gauge[k]; !ok {
		err = myErr.ErrNotExist
	}
	return
}

func (r *MemStorageRepo) GetCounter(ctx context.Context, k string) (v domain.Counter, err error) {
	var ok bool
	r.mc.RLock()
	defer r.mc.RUnlock()
	if v, ok = r.Counter[k]; !ok {
		err = myErr.ErrNotExist
	}
	return
}

func (r *MemStorageRepo) GetAllGauges(ctx context.Context) (domain.Gauges, error) {
	var err error
	return r.Gauge, err
}

func (r *MemStorageRepo) GetAllCounters(ctx context.Context) (domain.Counters, error) {
	var err error
	return r.Counter, err
}

func (r *MemStorageRepo) SetMetrics(ctx context.Context, metrics []domain.Metric) (newMetrics []domain.Metric, err error) {
	for _, metric := range metrics {
		switch metric.MType {
		case constant.MetricTypeGauge:
			if err = r.SetGauge(ctx, metric.ID, *metric.Value); err != nil {
				return
			}
			newMetrics = append(newMetrics, metric)
		case constant.MetricTypeCounter:
			var current domain.Counter
			if current, err = r.GetCounter(ctx, metric.ID); err != nil && !errors.Is(err, myErr.ErrNotExist) {
				return
			}
			delta := current + *metric.Delta
			if err = r.SetCounter(ctx, metric.ID, delta); err != nil {
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
