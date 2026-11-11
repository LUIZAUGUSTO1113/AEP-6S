package station

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct {
	collection *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{collection: db.Collection("stations")}
}

// EnsureIndexes creates the unique index on code. It is idempotent, so it is
// safe to call on every startup.
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "code", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("code_unique"),
	})
	if err != nil {
		return fmt.Errorf("failed to create stations code index: %w", err)
	}
	return nil
}

func (r *Repository) Create(ctx context.Context, station *Station) error {
	if station.ID.IsZero() {
		station.ID = bson.NewObjectID()
	}
	if _, err := r.collection.InsertOne(ctx, station); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrDuplicateCode
		}
		return fmt.Errorf("failed to insert station in mongodb: %w", err)
	}
	return nil
}

func (r *Repository) FindAll(ctx context.Context) ([]Station, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("error querying stations: %w", err)
	}
	defer cursor.Close(ctx)

	stations := []Station{}
	if err := cursor.All(ctx, &stations); err != nil {
		return nil, fmt.Errorf("error decoding stations: %w", err)
	}
	return stations, nil
}

func (r *Repository) FindByID(ctx context.Context, id bson.ObjectID) (*Station, error) {
	var station Station
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&station); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, mongo.ErrNoDocuments
		}
		return nil, fmt.Errorf("error querying station by id %s: %w", id.Hex(), err)
	}
	return &station, nil
}

// FindByIDForWrite takes a write lock in the transaction so concurrent station
// deletion and sample creation cannot commit an orphaned reference.
func (r *Repository) FindByIDForWrite(ctx context.Context, id bson.ObjectID) (*Station, error) {
	var station Station
	err := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": id},
		bson.M{"$inc": bson.M{"_sample_revision": 1}}).Decode(&station)
	if err != nil {
		return nil, err
	}
	return &station, nil
}

func (r *Repository) Update(ctx context.Context, id bson.ObjectID, station *Station) error {
	update := bson.M{"$set": bson.M{
		"name":       station.Name,
		"river":      station.River,
		"code":       station.Code,
		"location":   station.Location,
		"thresholds": station.Thresholds,
	}}
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrDuplicateCode
		}
		return fmt.Errorf("failed to update station in mongodb: %w", err)
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id bson.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete station in mongodb: %w", err)
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
