package repository

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/MrSwed/go-musthave-metrics/internal/config"
	myErr "github.com/MrSwed/go-musthave-metrics/internal/errors"
)

type MemStorage interface {
	SetGauge(k string, v float64) error
	SetCounter(k string, v int64) error
	GetGauge(k string) (float64, error)
	GetCounter(k string) (int64, error)
	GetAllCounters() (map[string]int64, error)
	GetAllGauges() (map[string]float64, error)
	Save() error
	restore() error
}

type MemStorageCounter struct {
	Counter map[string]int64 `json:"counter"`
	mc      sync.RWMutex
}

type MemStorageGauge struct {
	Gauge map[string]float64 `json:"gauge"`
	mg    sync.RWMutex
}

type MemStorageRepository struct {
	MemStorageCounter
	MemStorageGauge
	c *config.StorageConfig
}

func NewMemRepository(c *config.StorageConfig) (*MemStorageRepository, error) {
	m := &MemStorageRepository{
		c:                 c,
		MemStorageCounter: MemStorageCounter{Counter: map[string]int64{}},
		MemStorageGauge:   MemStorageGauge{Gauge: map[string]float64{}},
	}
	if c.FileStoragePath != "" {
		if c.StorageRestore {
			if err := m.restore(); err != nil {
				return nil, err
			}
		}
		if c.StoreInterval > 0 {
			go func() {
				for {
					time.Sleep(time.Duration(c.StoreInterval) * time.Second)
					_ = m.Save()
				}
			}()
		}
	}
	return m, nil
}

func (m *MemStorageRepository) SetGauge(k string, v float64) (err error) {
	m.mg.Lock()
	defer m.mg.Unlock()
	m.Gauge[k] = v
	if m.c.StoreInterval == 0 {
		err = m.Save()
	}
	return
}

func (m *MemStorageRepository) SetCounter(k string, v int64) (err error) {
	m.mc.Lock()
	defer m.mc.Unlock()
	m.Counter[k] = v
	if m.c.StoreInterval == 0 {
		err = m.Save()
	}
	return
}

func (m *MemStorageRepository) GetGauge(k string) (v float64, err error) {
	var ok bool
	m.mg.RLock()
	defer m.mg.RUnlock()
	if v, ok = m.Gauge[k]; !ok {
		err = myErr.ErrNotExist
	}
	return
}

func (m *MemStorageRepository) GetCounter(k string) (v int64, err error) {
	var ok bool
	m.mc.RLock()
	defer m.mc.RUnlock()
	if v, ok = m.Counter[k]; !ok {
		err = myErr.ErrNotExist
	}
	return
}

func (m *MemStorageRepository) GetAllGauges() (map[string]float64, error) {
	var err error
	return m.Gauge, err
}

func (m *MemStorageRepository) GetAllCounters() (map[string]int64, error) {
	var err error
	return m.Counter, err
}

func (m *MemStorageRepository) restore() (err error) {
	if m.c.FileStoragePath == "" {
		return nil
	}
	var data []byte
	if data, err = os.ReadFile(m.c.FileStoragePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &m)
}

func (m *MemStorageRepository) Save() error {
	if m.c.FileStoragePath == "" {
		return nil
	}
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(m.c.FileStoragePath, data, 0644)
}
