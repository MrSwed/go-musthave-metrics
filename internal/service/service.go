package service

import "github.com/MrSwed/go-musthave-metrics/internal/repository"

type Service struct {
	Metrics
}

func NewService(r repository.Repository) *Service {
	return &Service{Metrics: NewMemStorage(r)}
}
