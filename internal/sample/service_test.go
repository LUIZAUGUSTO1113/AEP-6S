package sample

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestServiceValidation(t *testing.T) {
	for _, tc := range []struct {
		name, parameter string
		stationID       *bson.ObjectID
		value           float64
		want            error
	}{
		{"station required", "pH", nil, 7, ErrInvalidStationID},
		{"parameter required", "", &testStationID, 7, ErrParameterRequired},
		{"negative pH", "pH", &testStationID, -0.01, ErrInvalidPH},
		{"high pH", "pH", &testStationID, 14.01, ErrInvalidPH},
		{"case and spaces", " PH ", &testStationID, 15, ErrInvalidPH},
		{"NaN", "pH", &testStationID, math.NaN(), ErrInvalidPH},
		{"infinity", "pH", &testStationID, math.Inf(1), ErrInvalidPH},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := testService(&repositoryStub{})
			s := &Sample{StationID: tc.stationID, Parameter: tc.parameter, Value: tc.value}
			require.ErrorIs(t, svc.CreateSample(context.Background(), s), tc.want)
			got, err := svc.UpdateSample(context.Background(), testSampleID.Hex(), s)
			require.ErrorIs(t, err, tc.want)
			require.Nil(t, got)
		})
	}
}

func TestServiceValidSamples(t *testing.T) {
	for _, tc := range []struct {
		name, parameter string
		value           float64
	}{
		{"minimum", "pH", 0}, {"maximum", "pH", 14}, {"normal", "pH", 7.2},
		{"other parameter", "Turbidity", 25},
	} {
		for _, suppliedDate := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/date=%v", tc.name, suppliedDate), func(t *testing.T) {
				for _, operation := range []string{"create", "update"} {
					s := &Sample{StationID: &testStationID, River: "ignored", Parameter: tc.parameter, Value: tc.value}
					date := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
					if suppliedDate {
						s.CollectedAt = date
					}
					before := time.Now()
					calls := 0
					check := func(ctx context.Context, got *Sample) error {
						calls++
						require.Same(t, s, got)
						require.Equal(t, "Rio Paraná", got.River)
						if suppliedDate {
							require.Equal(t, date, got.CollectedAt)
						} else {
							require.False(t, got.CollectedAt.Before(before))
							require.False(t, got.CollectedAt.After(time.Now()))
						}
						return nil
					}
					r := &repositoryStub{create: check,
						update: func(ctx context.Context, id string, got *Sample) error {
							require.Equal(t, testSampleID.Hex(), id)
							return check(ctx, got)
						},
						findByID: func(_ context.Context, id string) (*Sample, error) {
							require.Equal(t, testSampleID.Hex(), id)
							require.Equal(t, 1, calls)
							return s, nil
						},
					}
					svc := testService(r)
					if operation == "create" {
						require.NoError(t, svc.CreateSample(context.Background(), s))
					} else {
						got, err := svc.UpdateSample(context.Background(), testSampleID.Hex(), s)
						require.NoError(t, err)
						require.Same(t, s, got)
						require.Equal(t, "Rio Paraná", got.River)
					}
					require.Equal(t, 1, calls)
				}
			})
		}
	}
}

func TestServiceRepositoryErrors(t *testing.T) {
	failure := errors.New("storage unavailable")
	for _, tc := range []struct {
		name         string
		source, want error
	}{
		{"invalid ID", fmt.Errorf("wrapped: %w", errInvalidIDFromRepo), ErrInvalidID},
		{"not found", fmt.Errorf("wrapped: %w", mongo.ErrNoDocuments), ErrSampleNotFound},
		{"storage", failure, failure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := testService(&repositoryStub{
				delete: func(_ context.Context, id string) error { require.Equal(t, testSampleID.Hex(), id); return tc.source },
				update: func(context.Context, string, *Sample) error { return tc.source },
			})
			require.ErrorIs(t, svc.DeleteSample(context.Background(), testSampleID.Hex()), tc.want)
			got, err := svc.UpdateSample(context.Background(), testSampleID.Hex(), &Sample{StationID: &testStationID, Parameter: "pH", Value: 7})
			require.Nil(t, got)
			require.ErrorIs(t, err, tc.want)
		})
	}
	svc := testService(&repositoryStub{
		create:   func(context.Context, *Sample) error { return failure },
		delete:   func(context.Context, string) error { return nil },
		update:   func(context.Context, string, *Sample) error { return nil },
		findByID: func(context.Context, string) (*Sample, error) { return nil, failure },
	})
	s := &Sample{StationID: &testStationID, Parameter: "pH", Value: 7}
	require.ErrorIs(t, svc.CreateSample(context.Background(), s), failure)
	require.NoError(t, svc.DeleteSample(context.Background(), testSampleID.Hex()))
	got, err := svc.UpdateSample(context.Background(), testSampleID.Hex(), s)
	require.Nil(t, got)
	require.ErrorIs(t, err, failure)
}

func TestServiceList(t *testing.T) {
	for _, list := range [][]Sample{nil, {}, {{StationID: &testStationID, Parameter: "pH", Value: 7}}} {
		svc := testService(&repositoryStub{findAll: func(context.Context) ([]Sample, error) { return list, nil }})
		got, err := svc.GetAllSamples(context.Background(), "")
		require.NoError(t, err)
		require.Equal(t, list, got)
	}
	failure := errors.New("query failed")
	svc := testService(&repositoryStub{findAll: func(context.Context) ([]Sample, error) { return nil, failure }})
	got, err := svc.GetAllSamples(context.Background(), "")
	require.Nil(t, got)
	require.ErrorIs(t, err, failure)
}
