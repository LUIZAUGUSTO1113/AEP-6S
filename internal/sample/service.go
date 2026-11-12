package sample

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/LUIZAUGUSTO1113/AEP-6S/internal/alert"
	"github.com/LUIZAUGUSTO1113/AEP-6S/internal/database"
	"github.com/LUIZAUGUSTO1113/AEP-6S/internal/station"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrInvalidStationID  = errors.New("invalid station id format")
	ErrStationNotFound   = errors.New("station not found")
	ErrValueRequired     = errors.New("Value is required.")
	ErrParameterRequired = errors.New("Parameter is required.")
	ErrInvalidID         = errors.New("invalid sample id format")
	ErrSampleNotFound    = errors.New("sample not found")
	ErrInvalidPH         = errors.New("pH must be between 0 and 14 (inclusive).")
	errInvalidIDFromRepo = errors.New("invalid object id format")
)

type SampleRepository interface {
	Create(context.Context, *Sample) error
	FindAll(context.Context) ([]Sample, error)
	FindByStation(context.Context, bson.ObjectID) ([]Sample, error)
	Update(context.Context, string, *Sample) error
	FindByID(context.Context, string) (*Sample, error)
	Delete(context.Context, string) error
}

type StationRepository interface {
	FindByIDForWrite(context.Context, bson.ObjectID) (*station.Station, error)
}

type AlertRepository interface {
	Upsert(context.Context, alert.Alert) error
	DeleteBySample(context.Context, bson.ObjectID) error
}

type Service struct {
	repo        SampleRepository
	stations    StationRepository
	alerts      AlertRepository
	transaction database.Transaction
}

func NewService(repo SampleRepository, stations StationRepository, alerts AlertRepository, transaction database.Transaction) *Service {
	return &Service{repo: repo, stations: stations, alerts: alerts, transaction: transaction}
}

func (s *Service) CreateSample(ctx context.Context, sample *Sample) error {
	return s.save(ctx, sample, func(ctx context.Context) error {
		return s.repo.Create(ctx, sample)
	})
}

func (s *Service) GetAllSamples(ctx context.Context, stationID string) ([]Sample, error) {
	if stationID == "" {
		return s.repo.FindAll(ctx)
	}
	id, err := bson.ObjectIDFromHex(stationID)
	if err != nil {
		return nil, ErrInvalidStationID
	}
	return s.repo.FindByStation(ctx, id)
}

func (s *Service) UpdateSample(ctx context.Context, id string, sample *Sample) (*Sample, error) {
	if _, err := bson.ObjectIDFromHex(id); err != nil {
		return nil, ErrInvalidID
	}
	var updated *Sample
	err := s.save(ctx, sample, func(ctx context.Context) error {
		if err := s.repo.Update(ctx, id, sample); err != nil {
			return sampleError(err)
		}
		var err error
		updated, err = s.repo.FindByID(ctx, id)
		return sampleError(err)
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) DeleteSample(ctx context.Context, id string) error {
	sampleID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidID
	}
	return s.transaction(ctx, func(ctx context.Context) error {
		if err := s.repo.Delete(ctx, id); err != nil {
			return sampleError(err)
		}
		return s.alerts.DeleteBySample(ctx, sampleID)
	})
}

func (s *Service) save(ctx context.Context, sample *Sample, persist func(context.Context) error) error {
	if sample.StationID == nil {
		return ErrInvalidStationID
	}
	if strings.TrimSpace(sample.Parameter) == "" {
		return ErrParameterRequired
	}
	if strings.EqualFold(strings.TrimSpace(sample.Parameter), "pH") && !(sample.Value >= 0 && sample.Value <= 14) {
		return ErrInvalidPH
	}
	now := time.Now().UTC()
	if sample.CollectedAt.IsZero() {
		sample.CollectedAt = now
	}

	return s.transaction(ctx, func(ctx context.Context) error {
		station, err := s.stations.FindByIDForWrite(ctx, *sample.StationID)
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrStationNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to load station: %w", err)
		}
		sample.River = station.River
		if err := persist(ctx); err != nil {
			return err
		}
		violated, message := station.Evaluate(sample.Parameter, sample.Value)
		if !violated {
			return s.alerts.DeleteBySample(ctx, sample.ID)
		}
		return s.alerts.Upsert(ctx, alert.Alert{
			SampleID: sample.ID, StationID: station.ID, Parameter: sample.Parameter,
			Value: sample.Value, Message: message, CreatedAt: now,
		})
	})
}

func sampleError(err error) error {
	switch {
	case errors.Is(err, mongo.ErrNoDocuments):
		return ErrSampleNotFound
	case errors.Is(err, errInvalidIDFromRepo):
		return ErrInvalidID
	default:
		return err
	}
}
