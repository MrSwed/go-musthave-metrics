package storage

var Store = NewMemStorage()

type MemStore interface {
	SetGauge(k string, v float64)
	IncreaseCounter(k string, v int64)
	GetGauge(k string) float64
	GetCounter(k string) int64
}

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
	//...
}

func NewMemStorage() *MemStorage {
	return &MemStorage{gauge: map[string]float64{}, counter: map[string]int64{}}
}

func (m *MemStorage) SetGauge(k string, v float64) {
	m.gauge[k] = v
}

func (m *MemStorage) IncreaseCounter(k string, v int64) {
	if _, ok := m.counter[k]; ok {
		m.counter[k] = m.counter[k] + v
	} else {
		m.counter[k] = v
	}
}

func (m *MemStorage) GetGauge(k string) float64 {
	return m.gauge[k]
}

func (m *MemStorage) GetCounter(k string) int64 {
	return m.counter[k]
}
