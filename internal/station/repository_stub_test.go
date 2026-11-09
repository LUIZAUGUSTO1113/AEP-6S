package station

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Unconfigured operations panic so unexpected persistence calls fail the test.
type repositoryStub struct {
	create   func(context.Context, *Station) error
	findAll  func(context.Context) ([]Station, error)
	findByID func(context.Context, bson.ObjectID) (*Station, error)
	update   func(context.Context, bson.ObjectID, *Station) error
	delete   func(context.Context, bson.ObjectID) error
}

func (r *repositoryStub) Create(ctx context.Context, s *Station) error { return r.create(ctx, s) }

func (r *repositoryStub) FindAll(ctx context.Context) ([]Station, error) { return r.findAll(ctx) }

func (r *repositoryStub) FindByID(ctx context.Context, id bson.ObjectID) (*Station, error) {
	return r.findByID(ctx, id)
}

func (r *repositoryStub) Update(ctx context.Context, id bson.ObjectID, s *Station) error {
	return r.update(ctx, id, s)
}

func (r *repositoryStub) Delete(ctx context.Context, id bson.ObjectID) error {
	return r.delete(ctx, id)
}

type sampleCounterStub func(context.Context, bson.ObjectID) (int64, error)

func (f sampleCounterStub) CountByStation(ctx context.Context, id bson.ObjectID) (int64, error) {
	return f(ctx, id)
}

func directTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}
