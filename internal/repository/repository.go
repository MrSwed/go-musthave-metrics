package repository

import (
	"fmt"
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
	if db == nil {
		return nil, fmt.Errorf("need db connector")
	}
	return &RepositoryStorage{
		MemStorage:  NewMemRepository(),
		FileStorage: NewFileStorageRepository(c),
		DBStorage:   NewDBStorageRepository(db),
	}, nil
}
