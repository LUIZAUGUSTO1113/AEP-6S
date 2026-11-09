package sample

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type SampleRequest struct {
	StationID   string     `json:"station_id" binding:"required" example:"507f1f77bcf86cd799439011"`
	Parameter   string     `json:"parameter" binding:"required" example:"pH"`
	Value       *float64   `json:"value" binding:"required" example:"7.2"`
	CollectedAt *time.Time `json:"collected_at,omitempty" example:"2026-09-04T21:00:00Z"`
}

func (r SampleRequest) toSample() (*Sample, error) {
	stationID, err := bson.ObjectIDFromHex(r.StationID)
	if err != nil {
		return nil, ErrInvalidStationID
	}
	if r.Value == nil {
		return nil, ErrValueRequired
	}

	sample := &Sample{StationID: &stationID, Parameter: r.Parameter, Value: *r.Value}
	if r.CollectedAt != nil {
		sample.CollectedAt = *r.CollectedAt
	}
	return sample, nil
}
