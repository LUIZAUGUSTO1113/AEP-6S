package station

import "go.mongodb.org/mongo-driver/v2/bson"

// Station is the complex document of the stations collection: location and
// thresholds live and die with the station, so they are embedded.
type Station struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name       string        `bson:"name" json:"name"`
	River      string        `bson:"river" json:"river"`
	Code       string        `bson:"code" json:"code"`
	Location   Location      `bson:"location" json:"location"`
	Thresholds []Threshold   `bson:"thresholds" json:"thresholds"`
}

type Location struct {
	City  string  `bson:"city" json:"city"`
	State string  `bson:"state" json:"state"`
	Lat   float64 `bson:"lat" json:"lat"`
	Lng   float64 `bson:"lng" json:"lng"`
}

// Threshold bounds are pointers so an absent bound differs from a bound of 0.
type Threshold struct {
	Parameter string   `bson:"parameter" json:"parameter"`
	Min       *float64 `bson:"min,omitempty" json:"min,omitempty"`
	Max       *float64 `bson:"max,omitempty" json:"max,omitempty"`
}
