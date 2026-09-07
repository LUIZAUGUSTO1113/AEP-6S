package sample

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Repository struct {
	collection *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		collection: db.Collection("samples"),
	}
}

func (r *Repository) Create(ctx context.Context, sample *Sample) error {
	if sample.ID.IsZero() {
		sample.ID = bson.NewObjectID()
	}

	_, err := r.collection.InsertOne(ctx, sample)
	if err != nil {
		return fmt.Errorf("failed to insert document in mongodb: %w", err)
	}

	return nil
}

func (r *Repository) FindAll(ctx context.Context) ([]Sample, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("error querying samples: %w", err)
	}
	defer cursor.Close(ctx)

	var samples []Sample
	if err := cursor.All(ctx, &samples); err != nil {
		return nil, fmt.Errorf("error decoding samples: %w", err)
	}

	return samples, nil
}

func (r *Repository) FindByRiver(ctx context.Context, riverName string) ([]Sample, error) {
	filter := bson.M{"river": riverName}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("error querying samples by river %s: %w", riverName, err)
	}
	defer cursor.Close(ctx)

	var samples []Sample
	if err := cursor.All(ctx, &samples); err != nil {
		return nil, fmt.Errorf("error decoding samples by river %s: %w", riverName, err)
	}

	return samples, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return errInvalidIDFromRepo
	}

	filter := bson.M{"_id": objectID}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete document in mongodb: %w", err)
	}

	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
