package service

import (
	"context"

	"github.com/MrSwed/go-musthave-metrics/internal/server/repository"
)

type MetricsFile interface {
	SaveToFile(ctx context.Context) (int64, error)
	RestoreFromFile(ctx context.Context) (int64, error)
}

func (s *MetricsService) SaveToFile(ctx context.Context) (n int64, err error) {
	var m *repository.MemStorageRepo
	m, err = s.r.MemStore(ctx)
	if err == nil {
		err = s.r.SaveToFile(m)
		n = int64(len(m.MemStorageCounter.Counter) + len(m.MemStorageGauge.Gauge))
	}
	return
}

func (s *MetricsService) RestoreFromFile(ctx context.Context) (n int64, err error) {
	var m *repository.MemStorageRepo
	m, err = s.r.MemStore(ctx)
	if err == nil {
		err = s.r.RestoreFromFile(m)
		n = int64(len(m.MemStorageCounter.Counter) + len(m.MemStorageGauge.Gauge))
	}
	return
}
