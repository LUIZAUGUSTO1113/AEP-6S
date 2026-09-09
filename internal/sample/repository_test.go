package sample

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

func integrationRepository(t *testing.T) (*Repository, context.Context) {
	t.Helper()
	if testing.Short() {
		t.Skip("MongoDB integration: run without -short for full coverage")
	}
	uri := os.Getenv("MONGO_TEST_URI")
	require.NotEmpty(t, uri, "set MONGO_TEST_URI and start docker compose -f docker-compose.test.yml up -d --wait")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	require.NoError(t, err)
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		require.NoError(t, client.Disconnect(cleanup))
	})
	require.NoError(t, client.Ping(ctx, nil), "test MongoDB unavailable")
	// Only this generated database is ever dropped; application DB names are ignored.
	db := client.Database("aep_test_" + bson.NewObjectID().Hex())
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		require.NoError(t, db.Drop(cleanup))
	})
	return NewRepository(db), ctx
}

func TestRepositoryCRUD(t *testing.T) {
	r, ctx := integrationRepository(t)
	empty, err := r.FindAll(ctx)
	require.NoError(t, err)
	require.Empty(t, empty)
	date := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	a := &Sample{River: "Rio Paraná", Parameter: "pH", Value: 7.2, CollectedAt: date}
	require.NoError(t, r.Create(ctx, a))
	require.False(t, a.ID.IsZero())
	b := &Sample{ID: bson.NewObjectID(), River: "Rio Iguaçu", Parameter: "Turbidity", Value: 25, CollectedAt: date}
	originalID := b.ID
	require.NoError(t, r.Create(ctx, b))
	require.Equal(t, originalID, b.ID)
	got, err := r.FindByID(ctx, a.ID.Hex())
	require.NoError(t, err)
	require.Equal(t, a, got)
	all, err := r.FindAll(ctx)
	require.NoError(t, err)
	require.ElementsMatch(t, []Sample{*a, *b}, all)
	filtered, err := r.FindByRiver(ctx, a.River)
	require.NoError(t, err)
	require.Equal(t, []Sample{*a}, filtered)
	filtered, err = r.FindByRiver(ctx, "Unknown")
	require.NoError(t, err)
	require.Empty(t, filtered)
	updated := &Sample{River: "Rio atualizado", Parameter: "pH", Value: 8, CollectedAt: date.Add(time.Hour)}
	require.NoError(t, r.Update(ctx, a.ID.Hex(), updated))
	require.Equal(t, a.ID, updated.ID)
	got, err = r.FindByID(ctx, a.ID.Hex())
	require.NoError(t, err)
	require.Equal(t, updated, got)
	// Matching an unchanged document must succeed even when ModifiedCount is zero.
	require.NoError(t, r.Update(ctx, a.ID.Hex(), updated))
	require.NoError(t, r.Delete(ctx, a.ID.Hex()))
	_, err = r.FindByID(ctx, a.ID.Hex())
	require.ErrorIs(t, err, mongo.ErrNoDocuments)
	all, err = r.FindAll(ctx)
	require.NoError(t, err)
	require.Equal(t, []Sample{*b}, all)
	require.ErrorIs(t, r.Delete(ctx, a.ID.Hex()), mongo.ErrNoDocuments)
	require.ErrorIs(t, r.Update(ctx, a.ID.Hex(), updated), mongo.ErrNoDocuments)
	// Duplicate ID reaches the real database and preserves the driver error.
	err = r.Create(ctx, b)
	require.Error(t, err)
	require.True(t, mongo.IsDuplicateKeyError(err))
}

func TestRepositoryInvalidIDs(t *testing.T) {
	// Invalid IDs are rejected without contacting MongoDB.
	r := &Repository{}
	for _, id := range []string{"", "invalid", "123", "zzzzzzzzzzzzzzzzzzzzzzzz"} {
		t.Run(id, func(t *testing.T) {
			require.ErrorIs(t, r.Delete(context.Background(), id), errInvalidIDFromRepo)
			require.ErrorIs(t, r.Update(context.Background(), id, &Sample{}), errInvalidIDFromRepo)
			got, err := r.FindByID(context.Background(), id)
			require.Nil(t, got)
			require.ErrorIs(t, err, errInvalidIDFromRepo)
		})
	}
}

func TestRepositoryCancelledContext(t *testing.T) {
	r, _ := integrationRepository(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	id := bson.NewObjectID().Hex()
	require.ErrorIs(t, r.Create(ctx, &Sample{}), context.Canceled)
	_, err := r.FindAll(ctx)
	require.ErrorIs(t, err, context.Canceled)
	_, err = r.FindByRiver(ctx, "Rio")
	require.ErrorIs(t, err, context.Canceled)
	_, err = r.FindByID(ctx, id)
	require.ErrorIs(t, err, context.Canceled)
	require.ErrorIs(t, r.Update(ctx, id, &Sample{}), context.Canceled)
	require.ErrorIs(t, r.Delete(ctx, id), context.Canceled)
}

func TestRepositoryMalformedDocument(t *testing.T) {
	r, ctx := integrationRepository(t)
	id := bson.NewObjectID()
	_, err := r.collection.InsertOne(ctx, bson.M{"_id": id, "river": "Rio", "value": bson.M{"invalid": "number expected"}})
	require.NoError(t, err)
	_, err = r.FindAll(ctx)
	require.ErrorContains(t, err, "error decoding samples")
	_, err = r.FindByRiver(ctx, "Rio")
	require.ErrorContains(t, err, "error decoding samples by river")
	_, err = r.FindByID(ctx, id.Hex())
	require.ErrorContains(t, err, "error querying sample by id")
}
