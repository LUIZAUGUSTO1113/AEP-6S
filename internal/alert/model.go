package alert

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Alert struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	SampleID  bson.ObjectID `bson:"sample_id" json:"sample_id"`
	StationID bson.ObjectID `bson:"station_id" json:"station_id"`
	Parameter string        `bson:"parameter" json:"parameter"`
	Value     float64       `bson:"value" json:"value"`
	Message   string        `bson:"message" json:"message"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
}
