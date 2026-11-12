package station

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
	require.NoError(t, r.EnsureIndexes(ctx))
	require.NoError(t, r.EnsureIndexes(ctx), "index creation is idempotent")

	empty, err := r.FindAll(ctx)
	require.NoError(t, err)
	require.NotNil(t, empty, "empty list encodes as [] instead of null")
	require.Empty(t, empty)

	a := validStation()
	require.NoError(t, r.Create(ctx, a))
	require.False(t, a.ID.IsZero())
	b := validStation()
	b.ID, b.Code, b.Thresholds = bson.NewObjectID(), "PAR-001", []Threshold{}
	originalID := b.ID
	require.NoError(t, r.Create(ctx, b))
	require.Equal(t, originalID, b.ID)

	got, err := r.FindByID(ctx, a.ID)
	require.NoError(t, err)
	require.Equal(t, a, got, "nested location and thresholds round-trip, absent min stays nil")

	all, err := r.FindAll(ctx)
	require.NoError(t, err)
	require.ElementsMatch(t, []Station{*a, *b}, all)

	dup := validStation()
	require.ErrorIs(t, r.Create(ctx, dup), ErrDuplicateCode)

	changed := validStation()
	changed.Name, changed.Thresholds = "Novo nome", []Threshold{{Parameter: "pH", Min: ptr(6.5)}}
	require.NoError(t, r.Update(ctx, a.ID, changed))
	got, err = r.FindByID(ctx, a.ID)
	require.NoError(t, err)
	require.Equal(t, a.ID, got.ID)
	require.Equal(t, "Novo nome", got.Name)
	require.Equal(t, changed.Thresholds, got.Thresholds)

	changed.Code = b.Code
	require.ErrorIs(t, r.Update(ctx, a.ID, changed), ErrDuplicateCode)

	missing := bson.NewObjectID()
	_, err = r.FindByID(ctx, missing)
	require.ErrorIs(t, err, mongo.ErrNoDocuments)
	require.ErrorIs(t, r.Update(ctx, missing, validStation()), mongo.ErrNoDocuments)

	require.NoError(t, r.Delete(ctx, a.ID))
	require.ErrorIs(t, r.Delete(ctx, a.ID), mongo.ErrNoDocuments)
	_, err = r.FindByID(ctx, a.ID)
	require.ErrorIs(t, err, mongo.ErrNoDocuments)
}

func TestRepositoryErrorsWhenDatabaseUnavailable(t *testing.T) {
	r, ctx := integrationRepository(t)
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	id := bson.NewObjectID()

	require.Error(t, r.EnsureIndexes(canceled))
	err := r.Create(canceled, validStation())
	require.Error(t, err)
	require.NotErrorIs(t, err, ErrDuplicateCode)
	_, err = r.FindAll(canceled)
	require.Error(t, err)
	_, err = r.FindByID(canceled, id)
	require.Error(t, err)
	require.NotErrorIs(t, err, mongo.ErrNoDocuments)
	err = r.Update(canceled, id, validStation())
	require.Error(t, err)
	require.NotErrorIs(t, err, mongo.ErrNoDocuments)
	err = r.Delete(canceled, id)
	require.Error(t, err)
	require.NotErrorIs(t, err, mongo.ErrNoDocuments)
}
