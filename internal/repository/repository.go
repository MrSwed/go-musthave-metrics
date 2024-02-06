package repository

import (
	"github.com/MrSwed/go-musthave-metrics/internal/config"
	"github.com/jmoiron/sqlx"
)

//go:generate  mockgen -destination=../mock/repository/repository.go -package=mock "github.com/MrSwed/go-musthave-metrics/internal/repository" Repository

type Repository interface {
	MemStorage
	FileStorage
	DBStorage
}

type RepositoryStorage struct {
	MemStorage
	FileStorage
	DBStorage
}

func NewRepository(c *config.StorageConfig, db *sqlx.DB) (*RepositoryStorage, error) {
	return &RepositoryStorage{
		MemStorage:  NewMemRepository(),
		FileStorage: NewFileStorageRepository(c),
		DBStorage:   NewDBStorageRepository(db),
	}, nil
}
