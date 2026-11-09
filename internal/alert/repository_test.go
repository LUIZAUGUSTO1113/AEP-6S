package alert

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestRepositoryAlertLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("MongoDB integration: run without -short for full coverage")
	}
	uri := os.Getenv("MONGO_TEST_URI")
	require.NotEmpty(t, uri, "set MONGO_TEST_URI and start the isolated MongoDB replica set")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	require.NoError(t, err)
	db := client.Database("aep_test_" + bson.NewObjectID().Hex())
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		defer client.Disconnect(ctx)
		require.NoError(t, db.Drop(ctx))
	})

	repo := NewRepository(db)
	require.NoError(t, repo.EnsureIndexes(ctx))
	require.NoError(t, repo.EnsureIndexes(ctx))
	a := Alert{
		SampleID: bson.NewObjectID(), StationID: bson.NewObjectID(),
		Parameter: "pH", Value: 4.2, Message: "Abaixo do limiar",
		CreatedAt: time.Date(2026, 10, 27, 14, 0, 0, 0, time.UTC),
	}
	require.NoError(t, repo.Upsert(ctx, a))
	items, err := repo.Find(ctx, &a.StationID)
	require.NoError(t, err)
	require.Len(t, items, 1)
	original := items[0]
	require.False(t, original.ID.IsZero())
	require.Equal(t, a.SampleID, original.SampleID)
	require.Equal(t, a.Message, original.Message)

	oldStation := a.StationID
	a.StationID = bson.NewObjectID()
	a.Value = 10
	a.Message = "Acima do limiar"
	a.CreatedAt = a.CreatedAt.Add(time.Hour)
	require.NoError(t, repo.Upsert(ctx, a))
	items, err = repo.Find(ctx, &oldStation)
	require.NoError(t, err)
	require.Equal(t, []Alert{}, items)
	items, err = repo.Find(ctx, nil)
	require.NoError(t, err)
	require.Len(t, items, 1)
	a.ID = original.ID
	a.CreatedAt = original.CreatedAt
	require.Equal(t, a, items[0])

	_, err = repo.collection.InsertOne(ctx, bson.M{"sample_id": a.SampleID})
	require.True(t, mongo.IsDuplicateKeyError(err), "%v", err)
	require.NoError(t, repo.DeleteBySample(ctx, a.SampleID))
	require.NoError(t, repo.DeleteBySample(ctx, a.SampleID))
	items, err = repo.Find(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, []Alert{}, items)
}
