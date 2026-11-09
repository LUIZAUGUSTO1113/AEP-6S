package sample

import (
	"context"

	"github.com/LUIZAUGUSTO1113/AEP-6S/internal/alert"
	"github.com/LUIZAUGUSTO1113/AEP-6S/internal/station"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Unconfigured operations panic so unexpected persistence calls fail the test.
type repositoryStub struct {
	create        func(context.Context, *Sample) error
	findAll       func(context.Context) ([]Sample, error)
	findByStation func(context.Context, bson.ObjectID) ([]Sample, error)
	update        func(context.Context, string, *Sample) error
	findByID      func(context.Context, string) (*Sample, error)
	delete        func(context.Context, string) error
}

func (r *repositoryStub) Create(ctx context.Context, s *Sample) error   { return r.create(ctx, s) }
func (r *repositoryStub) FindAll(ctx context.Context) ([]Sample, error) { return r.findAll(ctx) }
func (r *repositoryStub) Update(ctx context.Context, id string, s *Sample) error {
	return r.update(ctx, id, s)
}
func (r *repositoryStub) FindByID(ctx context.Context, id string) (*Sample, error) {
	return r.findByID(ctx, id)
}
func (r *repositoryStub) Delete(ctx context.Context, id string) error { return r.delete(ctx, id) }
func (r *repositoryStub) FindByStation(ctx context.Context, id bson.ObjectID) ([]Sample, error) {
	return r.findByStation(ctx, id)
}

type stationLoader func(context.Context, bson.ObjectID) (*station.Station, error)

func (load stationLoader) FindByIDForWrite(ctx context.Context, id bson.ObjectID) (*station.Station, error) {
	return load(ctx, id)
}

type alertStub struct {
	upsert func(context.Context, alert.Alert) error
	remove func(context.Context, bson.ObjectID) error
}

func (a alertStub) Upsert(ctx context.Context, entry alert.Alert) error {
	return a.upsert(ctx, entry)
}

func (a alertStub) DeleteBySample(ctx context.Context, id bson.ObjectID) error {
	return a.remove(ctx, id)
}

var testStationID = bson.NewObjectID()
var testSampleID = bson.NewObjectID()

func directTransaction(ctx context.Context, run func(context.Context) error) error {
	return run(ctx)
}

func testService(repo SampleRepository) *Service {
	stations := stationLoader(func(_ context.Context, id bson.ObjectID) (*station.Station, error) {
		return &station.Station{ID: id, River: "Rio Paraná"}, nil
	})
	return NewService(repo, stations, alertStub{
		remove: func(context.Context, bson.ObjectID) error { return nil },
	}, directTransaction)
}
