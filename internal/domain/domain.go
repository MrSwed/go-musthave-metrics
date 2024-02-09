package domain

type Metric struct {
	ID    string   `json:"id" validate:"required"`
	MType string   `json:"type" validate:"required,oneof=gauge counter"`
	Delta *int64   `json:"delta,omitempty" validate:"required_if=MType counter,omitempty"`
	Value *float64 `json:"value,omitempty" validate:"required_if=MType gauge,omitempty"`
}
