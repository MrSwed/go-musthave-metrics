package domain

import "strconv"

type (
	Counter  int64
	Gauge    float64
	Gauges   map[string]Gauge
	Counters map[string]Counter
)

type Metric struct {
	ID    string   `json:"id" validate:"required"`
	MType string   `json:"type" validate:"required,oneof=gauge counter"`
	Delta *Counter `json:"delta,omitempty" validate:"required_if=MType counter,omitempty"`
	Value *Gauge   `json:"value,omitempty" validate:"required_if=MType gauge,omitempty"`
}

type ValidateMetrics struct {
	Metrics []Metric `validate:"required,gt=0,dive"`
}

func ParseGauge(str string) (v Gauge, err error) {
	var f float64
	if f, err = strconv.ParseFloat(str, 64); err == nil {
		v = Gauge(f)
	}
	return
}

func ParseCounter(str string) (v Counter, err error) {
	var f int64
	if f, err = strconv.ParseInt(str, 10, 64); err == nil {
		v = Counter(f)
	}
	return
}
