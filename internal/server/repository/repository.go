package repository

import (
	"context"

	"go-musthave-metrics/internal/server/config"
	"go-musthave-metrics/internal/server/domain"

	"github.com/jmoiron/sqlx"
)

// DataStorage methods
type DataStorage interface {
	// SetGauge save gauge to store
	SetGauge(ctx context.Context, k string, v domain.Gauge) error
	// SetCounter save counter to store
	SetCounter(ctx context.Context, k string, v domain.Counter) error
	// GetGauge get gauge from store
	GetGauge(ctx context.Context, k string) (domain.Gauge, error)
	// GetCounter get counter from store
	GetCounter(ctx context.Context, k string) (domain.Counter, error)
	// GetAllCounters get all counters from store
	GetAllCounters(ctx context.Context) (domain.Counters, error)
	// GetAllGauges get all gauges from store
	GetAllGauges(ctx context.Context) (domain.Gauges, error)
	// SetMetrics save several metrics to store
	SetMetrics(ctx context.Context, metrics []domain.Metric) ([]domain.Metric, error)
	Ping(ctx context.Context) error
	// MemStore return all metrics
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

// NewRepository return repository of database or memory if no db set
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
