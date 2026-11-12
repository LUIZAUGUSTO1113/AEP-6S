package station

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestGetStation(t *testing.T) {
	failure := errors.New("storage unavailable")
	id := bson.NewObjectID()
	found := &Station{ID: id, Code: "IVA-001"}

	for _, tc := range []struct {
		name    string
		id      string
		station *Station
		repoErr error
		queries bool
		want    error
	}{
		{name: "success", id: id.Hex(), station: found, queries: true},
		{name: "invalid id", id: "not-an-id", want: ErrInvalidID},
		{name: "not found", id: id.Hex(), repoErr: mongo.ErrNoDocuments, queries: true, want: ErrStationNotFound},
		{name: "repository failure", id: id.Hex(), repoErr: failure, queries: true, want: failure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			queried := false
			repo := &repositoryStub{findByID: func(_ context.Context, got bson.ObjectID) (*Station, error) {
				queried = true
				require.Equal(t, id, got)
				return tc.station, tc.repoErr
			}}

			got, err := NewService(repo, nil, directTransaction).GetStation(context.Background(), tc.id)

			require.Equal(t, tc.queries, queried)
			if tc.want != nil {
				require.ErrorIs(t, err, tc.want)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.Same(t, found, got)
		})
	}
}

func TestDeleteStation(t *testing.T) {
	failure := errors.New("storage unavailable")
	id := bson.NewObjectID()

	for _, tc := range []struct {
		name      string
		id        string
		count     int64
		countErr  error
		deleteErr error
		deletes   bool
		want      error
	}{
		{name: "success", id: id.Hex(), deletes: true},
		{name: "invalid id", id: "not-an-id", want: ErrInvalidID},
		{name: "blocked by samples", id: id.Hex(), count: 3, want: ErrStationHasSamples},
		{name: "count failure", id: id.Hex(), countErr: failure, want: failure},
		{name: "not found", id: id.Hex(), deleteErr: mongo.ErrNoDocuments, deletes: true, want: ErrStationNotFound},
		{name: "delete failure", id: id.Hex(), deleteErr: failure, deletes: true, want: failure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			deleted := false
			repo := &repositoryStub{delete: func(_ context.Context, got bson.ObjectID) error {
				deleted = true
				require.Equal(t, id, got)
				return tc.deleteErr
			}}
			counter := sampleCounterStub(func(_ context.Context, got bson.ObjectID) (int64, error) {
				require.Equal(t, id, got)
				return tc.count, tc.countErr
			})

			err := NewService(repo, counter, directTransaction).DeleteStation(context.Background(), tc.id)

			if tc.want == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.want)
			}
			require.Equal(t, tc.deletes, deleted)
		})
	}
}

func TestCreateStation(t *testing.T) {
	failure := errors.New("storage unavailable")
	for _, tc := range []struct {
		name    string
		input   *Station
		repoErr error
		creates bool
		want    error
	}{
		{name: "success", input: validStation(), creates: true},
		{name: "validation runs before storage", input: &Station{}, want: ErrNameRequired},
		{name: "duplicate code", input: validStation(), repoErr: ErrDuplicateCode, creates: true, want: ErrDuplicateCode},
		{name: "repository failure", input: validStation(), repoErr: failure, creates: true, want: failure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			created := false
			repo := &repositoryStub{create: func(_ context.Context, s *Station) error {
				created = true
				require.Same(t, tc.input, s)
				return tc.repoErr
			}}
			err := NewService(repo, nil, directTransaction).CreateStation(context.Background(), tc.input)
			require.Equal(t, tc.creates, created)
			if tc.want == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.want)
			}
		})
	}

	t.Run("normalizes before storing", func(t *testing.T) {
		s := validStation()
		s.Code, s.Location.State, s.Thresholds = " IVA-001 ", "pr", nil
		repo := &repositoryStub{create: func(context.Context, *Station) error { return nil }}
		require.NoError(t, NewService(repo, nil, directTransaction).CreateStation(context.Background(), s))
		require.Equal(t, "IVA-001", s.Code)
		require.Equal(t, "PR", s.Location.State)
		require.Equal(t, []Threshold{}, s.Thresholds)
	})
}

func TestListStations(t *testing.T) {
	stations := []Station{*validStation()}
	repo := &repositoryStub{findAll: func(context.Context) ([]Station, error) { return stations, nil }}
	got, err := NewService(repo, nil, directTransaction).ListStations(context.Background())
	require.NoError(t, err)
	require.Equal(t, stations, got)

	failure := errors.New("storage unavailable")
	repo = &repositoryStub{findAll: func(context.Context) ([]Station, error) { return nil, failure }}
	_, err = NewService(repo, nil, directTransaction).ListStations(context.Background())
	require.ErrorIs(t, err, failure)
}

func TestUpdateStation(t *testing.T) {
	failure := errors.New("storage unavailable")
	id := bson.NewObjectID()
	stored := &Station{ID: id, Code: "IVA-001"}

	for _, tc := range []struct {
		name      string
		id        string
		input     *Station
		updateErr error
		findErr   error
		updates   bool
		want      error
	}{
		{name: "success", id: id.Hex(), input: validStation(), updates: true},
		{name: "invalid id", id: "not-an-id", input: validStation(), want: ErrInvalidID},
		{name: "validation runs before storage", id: id.Hex(), input: &Station{Name: "x"}, want: ErrRiverRequired},
		{name: "not found", id: id.Hex(), input: validStation(), updateErr: mongo.ErrNoDocuments, updates: true, want: ErrStationNotFound},
		{name: "duplicate code", id: id.Hex(), input: validStation(), updateErr: ErrDuplicateCode, updates: true, want: ErrDuplicateCode},
		{name: "update failure", id: id.Hex(), input: validStation(), updateErr: failure, updates: true, want: failure},
		{name: "reload failure", id: id.Hex(), input: validStation(), findErr: failure, updates: true, want: failure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			updated := false
			repo := &repositoryStub{
				update: func(_ context.Context, got bson.ObjectID, s *Station) error {
					updated = true
					require.Equal(t, id, got)
					require.Same(t, tc.input, s)
					return tc.updateErr
				},
				findByID: func(_ context.Context, got bson.ObjectID) (*Station, error) {
					require.Equal(t, id, got)
					if tc.findErr != nil {
						return nil, tc.findErr
					}
					return stored, nil
				},
			}
			got, err := NewService(repo, nil, directTransaction).UpdateStation(context.Background(), tc.id, tc.input)
			require.Equal(t, tc.updates, updated)
			if tc.want != nil {
				require.ErrorIs(t, err, tc.want)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.Same(t, stored, got)
		})
	}
}
