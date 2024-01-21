package repository

import "github.com/MrSwed/go-musthave-metrics/internal/config"

func NewRepository(c *config.StorageConfig) MemStorage {
	return NewMemRepository(c)
}
