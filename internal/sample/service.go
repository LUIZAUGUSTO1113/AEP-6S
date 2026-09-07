package sample

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrRiverRequired     = errors.New("River is required.")
	ErrParameterRequired = errors.New("Parameter is required.")
	ErrInvalidID         = errors.New("invalid sample id format")
	ErrSampleNotFound    = mongo.ErrNoDocuments
	ErrInvalidPH         = errors.New("pH must be between 0 and 14 (inclusive).")
)

// SampleRepository defines the persistence operations used by Service.
type SampleRepository interface {
	Create(context.Context, *Sample) error
	FindAll(context.Context) ([]Sample, error)
	Update(context.Context, string, *Sample) error
	FindByID(context.Context, string) (*Sample, error)
	Delete(context.Context, string) error
}

type Service struct {
	repo SampleRepository
}

func NewService(repo SampleRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateSample(ctx context.Context, sample *Sample) error {
	if sample.River == "" {
		return ErrRiverRequired
	}

	if sample.Parameter == "" {
		return ErrParameterRequired
	}

	if err := validatePH(sample); err != nil {
		return err
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
		return nil, ErrRiverRequired
	}

	if sample.Parameter == "" {
		return nil, ErrParameterRequired
	}

	if err := validatePH(sample); err != nil {
		return nil, err
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

func validatePH(sample *Sample) error {
	if strings.EqualFold(strings.TrimSpace(sample.Parameter), "pH") && !(sample.Value >= 0 && sample.Value <= 14) {
		return ErrInvalidPH
	}
	return nil
}
