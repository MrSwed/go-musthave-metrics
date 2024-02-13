package service

import "github.com/MrSwed/go-musthave-metrics/internal/repository"

type MetricsFile interface {
	SaveToFile() (int64, error)
	RestoreFromFile() (int64, error)
}

func (s *MetricsService) SaveToFile() (n int64, err error) {
	var m *repository.MemStorageRepo
	m, err = s.r.MemStore()
	if err == nil {
		err = s.r.SaveToFile(m)
		n = int64(len(m.MemStorageCounter.Counter) + len(m.MemStorageGauge.Gauge))
	}
	return
}

func (s *MetricsService) RestoreFromFile() (n int64, err error) {
	var m *repository.MemStorageRepo
	m, err = s.r.MemStore()
	if err == nil {
		err = s.r.RestoreFromFile(m)
		n = int64(len(m.MemStorageCounter.Counter) + len(m.MemStorageGauge.Gauge))
	}
	return
}
