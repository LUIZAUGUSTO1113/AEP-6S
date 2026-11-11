package alert

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct {
	collection *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{collection: db.Collection("alerts")}
}

func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "sample_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "station_id", Value: 1}}},
	})
	return err
}

func (r *Repository) Upsert(ctx context.Context, alert Alert) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"sample_id": alert.SampleID}, bson.M{
		"$set": bson.M{
			"station_id": alert.StationID,
			"parameter":  alert.Parameter,
			"value":      alert.Value,
			"message":    alert.Message,
		},
		"$setOnInsert": bson.M{"created_at": alert.CreatedAt},
	}, options.UpdateOne().SetUpsert(true))
	return err
}

func (r *Repository) DeleteBySample(ctx context.Context, sampleID bson.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"sample_id": sampleID})
	return err
}

func (r *Repository) Find(ctx context.Context, stationID *bson.ObjectID) ([]Alert, error) {
	filter := bson.M{}
	if stationID != nil {
		filter["station_id"] = *stationID
	}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	alerts := []Alert{}
	if err := cursor.All(ctx, &alerts); err != nil {
		return nil, err
	}
	return alerts, nil
}
