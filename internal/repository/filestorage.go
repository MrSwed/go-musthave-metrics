package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/MrSwed/go-musthave-metrics/internal/config"
)

type FileStorage interface {
	SaveToFile(m *MemStorageRepo) error
	RestoreFromFile(m *MemStorageRepo) error
}

type FileStorageRepo struct {
	c *config.StorageConfig
}

func NewFileStorageRepository(c *config.StorageConfig) *FileStorageRepo {
	return &FileStorageRepo{
		c: c,
	}
}

func (f *FileStorageRepo) RestoreFromFile(m *MemStorageRepo) (err error) {
	if f.c.FileStoragePath == "" {
		return fmt.Errorf("no storage file provided")
	}
	var data []byte
	if data, err = os.ReadFile(f.c.FileStoragePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &m)
}

func (f *FileStorageRepo) SaveToFile(m *MemStorageRepo) error {
	if f.c.FileStoragePath == "" {
		return fmt.Errorf("no storage file provided")
	}
	m.mc.Lock()
	defer m.mc.Unlock()
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(f.c.FileStoragePath, data, 0644)
}
