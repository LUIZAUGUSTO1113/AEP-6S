package sample

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Sample struct {
	ID          bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	StationID   *bson.ObjectID `bson:"station_id,omitempty" json:"station_id,omitempty"`
	River       string         `bson:"river" json:"river"`
	Parameter   string         `bson:"parameter" json:"parameter"`
	Value       float64        `bson:"value" json:"value"`
	CollectedAt time.Time      `bson:"collected_at" json:"collected_at"`
}
