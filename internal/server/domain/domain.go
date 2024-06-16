package domain

import (
	"fmt"
	"strconv"
)

type (
	Counter  int64
	Gauge    float64
	Gauges   map[string]Gauge
	Counters map[string]Counter
)

// Metric common metric structure with validation
type Metric struct {
	Delta *Counter `json:"delta,omitempty" validate:"required_if=MType counter,omitempty"`
	Value *Gauge   `json:"value,omitempty" validate:"required_if=MType gauge,omitempty"`
	ID    string   `json:"id" validate:"required"`
	MType string   `json:"type" validate:"required,oneof=gauge counter"`
}

func (m Metric) String() (s string) {
	switch m.MType {
	case "counter":
		if m.Delta != nil {
			s = fmt.Sprintf("%v", *m.Delta)
		}
	case "gauge":
		if m.Value != nil {
			s = fmt.Sprintf("%v", *m.Value)
		}
	default:
	}
	return
}

// ValidateMetrics validate received metric collection data
type ValidateMetrics struct {
	Metrics []Metric `validate:"required,gt=0,dive"`
}

// ParseGauge parse gauge from string
func ParseGauge(str string) (v Gauge, err error) {
	var f float64
	if f, err = strconv.ParseFloat(str, 64); err == nil {
		v = Gauge(f)
	}
	return
}

// ParseCounter parse counter from string
func ParseCounter(str string) (v Counter, err error) {
	var f int64
	if f, err = strconv.ParseInt(str, 10, 64); err == nil {
		v = Counter(f)
	}
	return
}
