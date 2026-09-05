package sample

import (
	"context"
	"errors"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateSample(ctx context.Context, sample *Sample) error {
	if sample.River == "" {
		return errors.New("River is required.")
	}

	if sample.Parameter == "" {
		return errors.New("Parameter is required.")
	}

	if sample.CollectedAt.IsZero() {
		sample.CollectedAt = time.Now()
	}

	return s.repo.Create(ctx, sample)
}

func (s *Service) GetAllSamples(ctx context.Context) ([]Sample, error) {
	return s.repo.FindAll(ctx)
}
