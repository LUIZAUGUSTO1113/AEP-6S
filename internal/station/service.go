package station

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrInvalidID         = errors.New("invalid station id format")
	ErrStationNotFound   = mongo.ErrNoDocuments
	ErrStationHasSamples = errors.New("station still has samples and cannot be deleted")
	ErrDuplicateCode     = errors.New("a station with this code already exists")

	ErrNameRequired               = errors.New("Name is required.")
	ErrRiverRequired              = errors.New("River is required.")
	ErrCodeRequired               = errors.New("Code is required.")
	ErrCityRequired               = errors.New("Location city is required.")
	ErrStateRequired              = errors.New("Location state is required.")
	ErrThresholdParameterRequired = errors.New("Threshold parameter is required.")
	ErrThresholdRange             = errors.New("Threshold min must be less than or equal to max.")
	ErrInvalidPHThreshold         = errors.New("pH threshold must be between 0 and 14 (inclusive).")
)

// StationRepository defines the persistence operations used by Service.
type StationRepository interface {
	Create(context.Context, *Station) error
	FindAll(context.Context) ([]Station, error)
	FindByID(context.Context, bson.ObjectID) (*Station, error)
	Update(context.Context, bson.ObjectID, *Station) error
	Delete(context.Context, bson.ObjectID) error
}

// SampleCounter is implemented by the sample repository. It is declared here
// so station never imports sample; Go satisfies it structurally.
type SampleCounter interface {
	CountByStation(ctx context.Context, stationID bson.ObjectID) (int64, error)
}

type Service struct {
	repo    StationRepository
	samples SampleCounter
}

func NewService(repo StationRepository, samples SampleCounter) *Service {
	return &Service{repo: repo, samples: samples}
}

func (s *Service) CreateStation(ctx context.Context, station *Station) error {
	normalize(station)
	if err := validate(station); err != nil {
		return err
	}
	if err := s.repo.Create(ctx, station); err != nil {
		if errors.Is(err, ErrDuplicateCode) {
			return ErrDuplicateCode
		}
		return fmt.Errorf("service failed to create station: %w", err)
	}
	return nil
}

func (s *Service) ListStations(ctx context.Context) ([]Station, error) {
	stations, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("service failed to list stations: %w", err)
	}
	return stations, nil
}

// GetStation loads one station. The sample service uses it to validate
// station_id, copy the river and call Evaluate.
func (s *Service) GetStation(ctx context.Context, id string) (*Station, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidID
	}

	station, err := s.repo.FindByID(ctx, objectID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrStationNotFound
		}
		return nil, fmt.Errorf("service failed to get station: %w", err)
	}
	return station, nil
}

// UpdateStation replaces every editable field and returns the stored result.
func (s *Service) UpdateStation(ctx context.Context, id string, station *Station) (*Station, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidID
	}
	normalize(station)
	if err := validate(station); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, objectID, station); err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, ErrStationNotFound
		case errors.Is(err, ErrDuplicateCode):
			return nil, ErrDuplicateCode
		}
		return nil, fmt.Errorf("service failed to update station: %w", err)
	}

	updated, err := s.repo.FindByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("service failed to fetch updated station: %w", err)
	}
	return updated, nil
}

// DeleteStation removes a station only when no sample references it.
func (s *Service) DeleteStation(ctx context.Context, id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidID
	}

	count, err := s.samples.CountByStation(ctx, objectID)
	if err != nil {
		return fmt.Errorf("service failed to count station samples: %w", err)
	}
	if count > 0 {
		return ErrStationHasSamples
	}

	if err := s.repo.Delete(ctx, objectID); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrStationNotFound
		}
		return fmt.Errorf("service failed to delete station: %w", err)
	}
	return nil
}
