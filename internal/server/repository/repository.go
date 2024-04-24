package repository

import (
	"context"

	"github.com/MrSwed/go-musthave-metrics/internal/server/config"
	"github.com/MrSwed/go-musthave-metrics/internal/server/domain"
	"github.com/jmoiron/sqlx"
)

//go:generate  mockgen -destination=../mock/repository/repository.go -package=mock "github.com/MrSwed/go-musthave-metrics/internal/server/repository" Repository

type DataStorage interface {
	SetGauge(ctx context.Context, k string, v domain.Gauge) error
	SetCounter(ctx context.Context, k string, v domain.Counter) error
	GetGauge(ctx context.Context, k string) (domain.Gauge, error)
	GetCounter(ctx context.Context, k string) (domain.Counter, error)
	GetAllCounters(ctx context.Context) (domain.Counters, error)
	GetAllGauges(ctx context.Context) (domain.Gauges, error)
	SetMetrics(ctx context.Context, metrics []domain.Metric) ([]domain.Metric, error)
	Ping(ctx context.Context) error
	MemStore(ctx context.Context) (*MemStorageRepo, error)
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
