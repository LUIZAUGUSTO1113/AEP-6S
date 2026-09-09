package sample

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func controllerRouter(repo SampleRepository) http.Handler {
	c := NewController(NewService(repo))
	mux := http.NewServeMux()
	mux.HandleFunc("POST /samples", c.Create)
	mux.HandleFunc("GET /samples", c.GetAll)
	mux.HandleFunc("PUT /samples/{id}", c.Update)
	mux.HandleFunc("DELETE /samples/{id}", c.DeleteById)
	return mux
}

func TestControllerValidation(t *testing.T) {
	for _, method := range []string{"POST", "PUT"} {
		for _, tc := range []struct{ name, body, message string }{
			{"invalid JSON", "{", "Invalid JSON."},
			{"invalid date", `{"collected_at":"yesterday"}`, "Invalid JSON."},
			{"missing river", `{"parameter":"pH","value":7}`, ErrRiverRequired.Error()},
			{"missing parameter", `{"river":"Rio","value":7}`, ErrParameterRequired.Error()},
			{"invalid pH", `{"river":"Rio","parameter":"pH","value":15}`, ErrInvalidPH.Error()},
		} {
			t.Run(method+"/"+tc.name, func(t *testing.T) {
				path := "/samples"
				if method == "PUT" {
					path += "/id"
				}
				response := httptest.NewRecorder()
				controllerRouter(&repositoryStub{}).ServeHTTP(response, httptest.NewRequest(method, path, strings.NewReader(tc.body)))
				require.Equal(t, 400, response.Code)
				require.Equal(t, "application/json", response.Header().Get("Content-Type"))
				var body map[string]string
				require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
				require.Equal(t, tc.message, body["error"])
			})
		}
	}
}

func TestControllerWriteSuccess(t *testing.T) {
	for _, method := range []string{"POST", "PUT"} {
		for _, suppliedDate := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/date=%v", method, suppliedDate), func(t *testing.T) {
				id := bson.NewObjectID()
				var saved Sample
				save := func(_ context.Context, s *Sample) error { s.ID = id; saved = *s; return nil }
				repo := &repositoryStub{create: save,
					update: func(ctx context.Context, gotID string, s *Sample) error {
						require.Equal(t, id.Hex(), gotID)
						return save(ctx, s)
					},
					findByID: func(_ context.Context, gotID string) (*Sample, error) {
						require.Equal(t, id.Hex(), gotID)
						return &saved, nil
					},
				}
				payload := map[string]any{"river": "Rio Paraná", "parameter": "pH", "value": 7.2}
				date := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
				if suppliedDate {
					payload["collected_at"] = date
				}
				encoded, err := json.Marshal(payload)
				require.NoError(t, err)
				path, status := "/samples", 201
				if method == "PUT" {
					path += "/" + id.Hex()
					status = 200
				}
				response := httptest.NewRecorder()
				controllerRouter(repo).ServeHTTP(response, httptest.NewRequest(method, path, strings.NewReader(string(encoded))))
				require.Equal(t, status, response.Code)
				var got Sample
				require.NoError(t, json.Unmarshal(response.Body.Bytes(), &got))
				require.Equal(t, id, got.ID)
				require.Equal(t, "Rio Paraná", got.River)
				require.Equal(t, "pH", got.Parameter)
				require.Equal(t, 7.2, got.Value)
				require.False(t, got.CollectedAt.IsZero())
				if suppliedDate {
					require.True(t, date.Equal(got.CollectedAt))
				}
			})
		}
	}
}

func TestControllerRepositoryErrors(t *testing.T) {
	failure := errors.New("storage unavailable")
	for _, tc := range []struct {
		name, method string
		source       error
		status       int
	}{
		{"create failure", "POST", failure, 500},
		{"wrapped validation", "POST", fmt.Errorf("wrapped: %w", ErrInvalidPH), 400},
		{"update invalid ID", "PUT", errInvalidIDFromRepo, 400},
		{"update missing", "PUT", ErrSampleNotFound, 404},
		{"update failure", "PUT", failure, 500},
		{"delete invalid ID", "DELETE", errInvalidIDFromRepo, 400},
		{"delete missing", "DELETE", ErrSampleNotFound, 404},
		{"delete failure", "DELETE", failure, 500},
		{"list failure", "GET", failure, 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &repositoryStub{
				create:  func(context.Context, *Sample) error { return tc.source },
				update:  func(context.Context, string, *Sample) error { return tc.source },
				delete:  func(context.Context, string) error { return tc.source },
				findAll: func(context.Context) ([]Sample, error) { return nil, tc.source },
			}
			path := "/samples"
			if tc.method == "PUT" || tc.method == "DELETE" {
				path += "/id"
			}
			response := httptest.NewRecorder()
			controllerRouter(repo).ServeHTTP(response, httptest.NewRequest(tc.method, path, strings.NewReader(`{"river":"Rio","parameter":"pH","value":7}`)))
			require.Equal(t, tc.status, response.Code)
			var body map[string]string
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
			require.NotEmpty(t, body["error"])
		})
	}
}

func TestControllerListAndDelete(t *testing.T) {
	for _, samples := range [][]Sample{{}, {{River: "Rio", Parameter: "pH", Value: 7}}} {
		repo := &repositoryStub{findAll: func(context.Context) ([]Sample, error) { return samples, nil }}
		response := httptest.NewRecorder()
		controllerRouter(repo).ServeHTTP(response, httptest.NewRequest("GET", "/samples", nil))
		require.Equal(t, 200, response.Code)
		var got []Sample
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &got))
		require.Equal(t, samples, got)
	}
	called := false
	repo := &repositoryStub{delete: func(_ context.Context, id string) error { called = true; require.Equal(t, "id", id); return nil }}
	response := httptest.NewRecorder()
	controllerRouter(repo).ServeHTTP(response, httptest.NewRequest("DELETE", "/samples/id", nil))
	require.True(t, called)
	require.Equal(t, 204, response.Code)
	require.Empty(t, response.Body.String())
}
