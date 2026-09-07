package sample

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrInvalidID      = errors.New("invalid sample id format")
	ErrSampleNotFound = mongo.ErrNoDocuments
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

func (s *Service) DeleteSample(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrSampleNotFound
		}
		if errors.Is(err, errInvalidIDFromRepo) {
			return ErrInvalidID
		}
		return fmt.Errorf("service failed to delete sample: %w", err)
	}
	return nil
}

func (s *Service) UpdateSample(ctx context.Context, id string, sample *Sample) (*Sample, error) {
	if sample.River == "" {
		return nil, errors.New("River is required.")
	}

	if sample.Parameter == "" {
		return nil, errors.New("Parameter is required.")
	}

	if sample.CollectedAt.IsZero() {
		sample.CollectedAt = time.Now()
	}

	if err := s.repo.Update(ctx, id, sample); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrSampleNotFound
		}
		if errors.Is(err, errInvalidIDFromRepo) {
			return nil, ErrInvalidID
		}
		return nil, fmt.Errorf("service failed to update sample: %w", err)
	}

	updated, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service failed to fetch updated sample: %w", err)
	}

	return updated, nil
}

var errInvalidIDFromRepo = errors.New("invalid object id format")
