package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"go-musthave-metrics/internal/server/config"
)

// FileStorage handle file storage methods
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

// RestoreFromFile restore storage from file
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

// SaveToFile save storage to file
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
