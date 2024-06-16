package repository

import (
	"context"
	"errors"
	"sync"

	"go-musthave-metrics/internal/server/constant"
	"go-musthave-metrics/internal/server/domain"
	myErr "go-musthave-metrics/internal/server/errors"
)

// MemStorageCounter is counter storage
type MemStorageCounter struct {
	Counter domain.Counters `json:"counter"`
	mc      sync.RWMutex
}

// MemStorageGauge is gauge store
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

// Ping for memory storage it always true
func (r *MemStorageRepo) Ping(_ context.Context) (err error) {
	return
}

// MemStore return memory store off all metrics
func (r *MemStorageRepo) MemStore(_ context.Context) (*MemStorageRepo, error) {
	return r, nil
}

// SetGauge save gauge to memory store
func (r *MemStorageRepo) SetGauge(_ context.Context, k string, v domain.Gauge) (err error) {
	r.mg.Lock()
	defer r.mg.Unlock()
	r.Gauge[k] = v
	return
}

// SetCounter save counter cot memory store
func (r *MemStorageRepo) SetCounter(_ context.Context, k string, v domain.Counter) (err error) {
	r.mc.Lock()
	defer r.mc.Unlock()
	r.Counter[k] = v
	return
}

// GetGauge get gauge from memory store
func (r *MemStorageRepo) GetGauge(_ context.Context, k string) (v domain.Gauge, err error) {
	var ok bool
	r.mg.RLock()
	defer r.mg.RUnlock()
	if v, ok = r.Gauge[k]; !ok {
		err = myErr.ErrNotExist
	}
	return
}

// GetCounter get counter from memory store
func (r *MemStorageRepo) GetCounter(_ context.Context, k string) (v domain.Counter, err error) {
	var ok bool
	r.mc.RLock()
	defer r.mc.RUnlock()
	if v, ok = r.Counter[k]; !ok {
		err = myErr.ErrNotExist
	}
	return
}

// GetAllGauges get all gauges from memory store
func (r *MemStorageRepo) GetAllGauges(_ context.Context) (domain.Gauges, error) {
	var err error
	return r.Gauge, err
}

// GetAllCounters get all counter from memory store
func (r *MemStorageRepo) GetAllCounters(_ context.Context) (domain.Counters, error) {
	var err error
	return r.Counter, err
}

// SetMetrics save several metrics to memory store
func (r *MemStorageRepo) SetMetrics(ctx context.Context, metrics []domain.Metric) (newMetrics []domain.Metric, err error) {
	newMetrics = make([]domain.Metric, len(metrics))
	for i, metric := range metrics {
		switch metric.MType {
		case constant.MetricTypeGauge:
			if err = r.SetGauge(ctx, metric.ID, *metric.Value); err != nil {
				return
			}
		case constant.MetricTypeCounter:
			var current domain.Counter
			if current, err = r.GetCounter(ctx, metric.ID); err != nil && !errors.Is(err, myErr.ErrNotExist) {
				return
			}
			*metric.Delta += current
			if err = r.SetCounter(ctx, metric.ID, *metric.Delta); err != nil {
				return
			}
		}
		newMetrics[i] = metric
	}
	return
}
