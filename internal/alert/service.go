package alert

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrInvalidStationID = errors.New("invalid station id format")

type AlertRepository interface {
	Find(context.Context, *bson.ObjectID) ([]Alert, error)
}

type Service struct {
	repo AlertRepository
}

func NewService(repo AlertRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, stationID string) ([]Alert, error) {
	if stationID == "" {
		return s.repo.Find(ctx, nil)
	}
	id, err := bson.ObjectIDFromHex(stationID)
	if err != nil {
		return nil, ErrInvalidStationID
	}
	return s.repo.Find(ctx, &id)
}
