package sample

import "context"

// Unconfigured operations panic so unexpected persistence calls fail the test.
type repositoryStub struct {
	create   func(context.Context, *Sample) error
	findAll  func(context.Context) ([]Sample, error)
	update   func(context.Context, string, *Sample) error
	findByID func(context.Context, string) (*Sample, error)
	delete   func(context.Context, string) error
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
