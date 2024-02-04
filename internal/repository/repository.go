package repository

import "github.com/MrSwed/go-musthave-metrics/internal/config"

//go:generate  mockgen -destination=../mock/repository/repository.go -package=mock "github.com/MrSwed/go-musthave-metrics/internal/repository" Repository

type Repository interface {
	MemStorage
	FileStorage
}

type RepositoryStorage struct {
	MemStorage
	FileStorage
}

func NewRepository(c *config.StorageConfig) RepositoryStorage {
	return RepositoryStorage{
		MemStorage:  NewMemRepository(),
		FileStorage: NewFileStorageRepository(c),
	}
}
