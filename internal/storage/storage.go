package storage

var Store = NewMemStorage()

type MemStore interface {
	Gauge(k string, v float64)
	Counter(k string, v int64)
	Get(k string) interface{}
}

type MemStorage struct {
	store map[string]float64
	//...
}

func NewMemStorage() *MemStorage {
	return &MemStorage{store: map[string]float64{}}
}

func (m *MemStorage) Gauge(k string, v float64) {
	m.store[k] = v
}

func (m *MemStorage) Counter(k string, v int64) {
	if _, ok := m.store[k]; ok {
		m.store[k] = m.store[k] + float64(v)
	} else {
		// todo: Не понятно, долно ли новое создаваться ?
		m.store[k] = float64(v)
	}
}

func (m *MemStorage) Get(k string) interface{} {
	return m.store[k]
}
