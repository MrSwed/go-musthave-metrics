package repository

import (
	"github.com/MrSwed/go-musthave-metrics/internal/config"
	"github.com/jmoiron/sqlx"
)

//go:generate  mockgen -destination=../mock/repository/repository.go -package=mock "github.com/MrSwed/go-musthave-metrics/internal/repository" Repository

type DataStorage interface {
	SetGauge(k string, v float64) error
	SetCounter(k string, v int64) error
	GetGauge(k string) (float64, error)
	GetCounter(k string) (int64, error)
	GetAllCounters() (map[string]int64, error)
	GetAllGauges() (map[string]float64, error)
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
