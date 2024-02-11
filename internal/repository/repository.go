package repository

import (
	"github.com/MrSwed/go-musthave-metrics/internal/config"
	"github.com/MrSwed/go-musthave-metrics/internal/domain"
	"github.com/jmoiron/sqlx"
)

//go:generate  mockgen -destination=../mock/repository/repository.go -package=mock "github.com/MrSwed/go-musthave-metrics/internal/repository" Repository

type DataStorage interface {
	SetGauge(k string, v domain.Gauge) error
	SetCounter(k string, v domain.Counter) error
	GetGauge(k string) (domain.Gauge, error)
	GetCounter(k string) (domain.Counter, error)
	GetAllCounters() (domain.Counters, error)
	GetAllGauges() (domain.Gauges, error)
	SetMetrics(metrics []domain.Metric) ([]domain.Metric, error)
}

type Repository interface {
	DataStorage
	FileStorage
}

type Storage struct {
	DataStorage
	FileStorage
}

func NewRepository(c *config.StorageConfig, db *sqlx.DB) (s Storage) {
	if db != nil {
		s = Storage{
			DataStorage: NewDBStorageRepository(db),
			FileStorage: NewFileStorageRepository(c),
		}
	} else {
		s = Storage{
			DataStorage: NewMemRepository(),
			FileStorage: NewFileStorageRepository(c),
		}
	}
	return
}
