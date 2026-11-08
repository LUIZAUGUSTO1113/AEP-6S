package station

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func controllerRouter(repo StationRepository, counter SampleCounter) http.Handler {
	c := NewController(NewService(repo, counter))
	mux := http.NewServeMux()
	mux.HandleFunc("POST /stations", c.Create)
	mux.HandleFunc("GET /stations", c.GetAll)
	mux.HandleFunc("GET /stations/{id}", c.GetByID)
	mux.HandleFunc("PUT /stations/{id}", c.Update)
	mux.HandleFunc("DELETE /stations/{id}", c.DeleteByID)
	return mux
}

const validBody = `{
	"name": "Ponto Ponte do Rio Ivaí",
	"river": "Rio Ivaí",
	"code": "IVA-001",
	"location": {"city": "Maringá", "state": "pr", "lat": -23.42, "lng": -51.93},
	"thresholds": [
		{"parameter": "pH", "min": 6.0, "max": 9.0},
		{"parameter": "Turbidity", "max": 100},
		{"parameter": "Zero Floor", "min": 0}
	]
}`

func serve(t *testing.T, h http.Handler, method, path, body string) (*httptest.ResponseRecorder, map[string]string) {
	t.Helper()
	response := httptest.NewRecorder()
	h.ServeHTTP(response, httptest.NewRequest(method, path, strings.NewReader(body)))
	require.Equal(t, "application/json", response.Header().Get("Content-Type"))
	var errBody map[string]string
	if response.Code >= 400 {
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &errBody))
		require.NotEmpty(t, errBody["error"])
	}
	return response, errBody
}

func TestControllerValidation(t *testing.T) {
	id := bson.NewObjectID().Hex()
	for _, method := range []string{"POST", "PUT"} {
		for _, tc := range []struct{ name, body, message string }{
			{"invalid JSON", "{", "Invalid JSON."},
			{"wrong type", `{"name": 1}`, "Invalid JSON."},
			{"missing name", `{"river":"R","code":"C","location":{"city":"X","state":"PR"}}`, ErrNameRequired.Error()},
			{"missing state", `{"name":"N","river":"R","code":"C","location":{"city":"X"}}`, ErrStateRequired.Error()},
			{"min above max", `{"name":"N","river":"R","code":"C","location":{"city":"X","state":"PR"},"thresholds":[{"parameter":"Turbidity","min":10,"max":5}]}`, "thresholds[0]: " + ErrThresholdRange.Error()},
			{"pH threshold out of range", `{"name":"N","river":"R","code":"C","location":{"city":"X","state":"PR"},"thresholds":[{"parameter":"pH","max":15}]}`, "thresholds[0]: " + ErrInvalidPHThreshold.Error()},
		} {
			t.Run(method+"/"+tc.name, func(t *testing.T) {
				path := "/stations"
				if method == "PUT" {
					path += "/" + id
				}
				response, body := serve(t, controllerRouter(&repositoryStub{}, nil), method, path, tc.body)
				require.Equal(t, 400, response.Code)
				require.Equal(t, tc.message, body["error"])
			})
		}
	}
}

func TestControllerCreate(t *testing.T) {
	id := bson.NewObjectID()
	var saved Station
	repo := &repositoryStub{create: func(_ context.Context, s *Station) error {
		s.ID = id
		saved = *s
		return nil
	}}

	response, _ := serve(t, controllerRouter(repo, nil), "POST", "/stations", validBody)

	require.Equal(t, 201, response.Code)
	var got Station
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &got))
	require.Equal(t, saved, got)
	require.Equal(t, id, got.ID)
	require.Equal(t, "PR", got.Location.State, "state is normalized")
	require.Len(t, got.Thresholds, 3)
	require.Nil(t, got.Thresholds[1].Min, "absent bound stays absent")
	require.Equal(t, 0.0, *got.Thresholds[2].Min, "zero bound is kept")

	raw := map[string]any{}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &raw))
	require.NotContains(t, raw["thresholds"].([]any)[1], "min")
}

func TestControllerCreateWithoutThresholdsReturnsEmptyList(t *testing.T) {
	repo := &repositoryStub{create: func(context.Context, *Station) error { return nil }}
	body := `{"name":"N","river":"R","code":"C","location":{"city":"X","state":"PR"}}`
	response, _ := serve(t, controllerRouter(repo, nil), "POST", "/stations", body)
	require.Equal(t, 201, response.Code)
	require.Contains(t, response.Body.String(), `"thresholds":[]`)
}

func TestControllerReadAndUpdate(t *testing.T) {
	id := bson.NewObjectID()
	stored := validStation()
	stored.ID = id
	repo := &repositoryStub{
		findAll: func(context.Context) ([]Station, error) { return []Station{*stored}, nil },
		findByID: func(_ context.Context, got bson.ObjectID) (*Station, error) {
			require.Equal(t, id, got)
			return stored, nil
		},
		update: func(_ context.Context, got bson.ObjectID, s *Station) error {
			require.Equal(t, id, got)
			require.Equal(t, "IVA-001", s.Code)
			return nil
		},
	}
	h := controllerRouter(repo, nil)

	response, _ := serve(t, h, "GET", "/stations", "")
	require.Equal(t, 200, response.Code)
	var list []Station
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &list))
	require.Equal(t, []Station{*stored}, list)

	for _, tc := range []struct{ method, body string }{{"GET", ""}, {"PUT", validBody}} {
		response, _ = serve(t, h, tc.method, "/stations/"+id.Hex(), tc.body)
		require.Equal(t, 200, response.Code, tc.method)
		var got Station
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &got))
		require.Equal(t, *stored, got, tc.method)
	}
}

func TestControllerDelete(t *testing.T) {
	id := bson.NewObjectID()
	deleted := false
	repo := &repositoryStub{delete: func(_ context.Context, got bson.ObjectID) error {
		deleted = true
		require.Equal(t, id, got)
		return nil
	}}
	counter := sampleCounterStub(func(context.Context, bson.ObjectID) (int64, error) { return 0, nil })

	response, _ := serve(t, controllerRouter(repo, counter), "DELETE", "/stations/"+id.Hex(), "")

	require.True(t, deleted)
	require.Equal(t, 204, response.Code)
	require.Empty(t, response.Body.String())
}

func TestControllerErrorStatuses(t *testing.T) {
	failure := errors.New("storage unavailable")
	id := bson.NewObjectID().Hex()
	for _, tc := range []struct {
		name, method, path string
		repoErr            error
		samples            int64
		status             int
	}{
		{"get invalid id", "GET", "/stations/bad", nil, 0, 400},
		{"put invalid id", "PUT", "/stations/bad", nil, 0, 400},
		{"delete invalid id", "DELETE", "/stations/bad", nil, 0, 400},
		{"get missing", "GET", "/stations/" + id, mongo.ErrNoDocuments, 0, 404},
		{"put missing", "PUT", "/stations/" + id, mongo.ErrNoDocuments, 0, 404},
		{"delete missing", "DELETE", "/stations/" + id, mongo.ErrNoDocuments, 0, 404},
		{"delete with samples", "DELETE", "/stations/" + id, nil, 2, 409},
		{"create duplicate code", "POST", "/stations", ErrDuplicateCode, 0, 409},
		{"put duplicate code", "PUT", "/stations/" + id, ErrDuplicateCode, 0, 409},
		{"create failure", "POST", "/stations", failure, 0, 500},
		{"list failure", "GET", "/stations", failure, 0, 500},
		{"get failure", "GET", "/stations/" + id, failure, 0, 500},
		{"put failure", "PUT", "/stations/" + id, failure, 0, 500},
		{"delete failure", "DELETE", "/stations/" + id, failure, 0, 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &repositoryStub{
				create:   func(context.Context, *Station) error { return tc.repoErr },
				findAll:  func(context.Context) ([]Station, error) { return nil, tc.repoErr },
				findByID: func(context.Context, bson.ObjectID) (*Station, error) { return nil, tc.repoErr },
				update:   func(context.Context, bson.ObjectID, *Station) error { return tc.repoErr },
				delete:   func(context.Context, bson.ObjectID) error { return tc.repoErr },
			}
			counter := sampleCounterStub(func(context.Context, bson.ObjectID) (int64, error) { return tc.samples, nil })
			body := ""
			if tc.method == "POST" || tc.method == "PUT" {
				body = validBody
			}
			response, _ := serve(t, controllerRouter(repo, counter), tc.method, tc.path, body)
			require.Equal(t, tc.status, response.Code)
		})
	}
}

func TestStatusForWrappedErrors(t *testing.T) {
	require.Equal(t, 400, statusFor(fmt.Errorf("thresholds[3]: %w", ErrThresholdRange)))
	require.Equal(t, 409, statusFor(fmt.Errorf("wrapped: %w", ErrStationHasSamples)))
	require.Equal(t, 500, statusFor(errors.New("other")))
}
