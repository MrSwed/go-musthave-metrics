package repository

type MemStorage interface {
	SetGauge(k string, v float64) error
	SetCounter(k string, v int64) error
	GetGauge(k string) (float64, error)
	GetCounter(k string) (int64, error)
}

type MemStorageRepository struct {
	gauge   map[string]float64
	counter map[string]int64
}

func NewMemRepository() *MemStorageRepository {
	return &MemStorageRepository{gauge: map[string]float64{}, counter: map[string]int64{}}
}

func (m *MemStorageRepository) SetGauge(k string, v float64) (err error) {
	m.gauge[k] = v
	return
}

func (m *MemStorageRepository) SetCounter(k string, v int64) (err error) {
	m.counter[k] = v
	return
}

func (m *MemStorageRepository) GetGauge(k string) (v float64, err error) {
	v = m.gauge[k]
	return
}

func (m *MemStorageRepository) GetCounter(k string) (v int64, err error) {
	v = m.counter[k]
	return
}
