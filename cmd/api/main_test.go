package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/LUIZAUGUSTO1113/AEP-6S/internal/alert"
	"github.com/LUIZAUGUSTO1113/AEP-6S/internal/database"
	"github.com/LUIZAUGUSTO1113/AEP-6S/internal/sample"
	"github.com/LUIZAUGUSTO1113/AEP-6S/internal/station"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type sampleCounterFunc func(context.Context, bson.ObjectID) (int64, error)

func (count sampleCounterFunc) CountByStation(ctx context.Context, id bson.ObjectID) (int64, error) {
	return count(ctx, id)
}

// This package runs in its own go test process. The server started by main
// terminates with that process; no application source changes are needed.
func TestMainHTTPIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("API startup integration requires test MongoDB")
	}
	uri := os.Getenv("MONGO_TEST_URI")
	require.NotEmpty(t, uri, "set MONGO_TEST_URI and start the test MongoDB")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	require.NoError(t, err)
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		require.NoError(t, client.Disconnect(cleanup))
	})
	require.NoError(t, client.Ping(ctx, nil))
	dbName := "aep_test_" + bson.NewObjectID().Hex()
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		require.NoError(t, client.Database(dbName).Drop(cleanup))
	})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())
	t.Setenv("MONGO_URI", uri)
	t.Setenv("MONGO_INITDB_DATABASE", dbName)
	t.Setenv("PORT", fmt.Sprint(port))
	go main()
	base := fmt.Sprintf("http://127.0.0.1:%d", port)
	httpClient := &http.Client{Timeout: 5 * time.Second}
	defer httpClient.CloseIdleConnections()
	require.Eventually(t, func() bool {
		response, err := httpClient.Get(base + "/samples")
		if err != nil {
			return false
		}
		defer response.Body.Close()
		return response.StatusCode == http.StatusOK
	}, 10*time.Second, 50*time.Millisecond, "API did not become ready")
	missing := bson.NewObjectID().Hex()
	sampleBody := func(stationID string, value float64) string {
		return fmt.Sprintf(`{"station_id":%q,"parameter":"pH","value":%g,"river":"ignored"}`, stationID, value)
	}
	for _, tc := range []struct {
		method, path, body string
		status             int
	}{
		{"GET", "/swagger/index.html", "", 200},
		{"POST", "/samples", sampleBody(missing, 15), 400},
		{"POST", "/samples", sampleBody(missing, 7), 404},
		{"PUT", "/samples/invalid", sampleBody(missing, 7), 400},
		{"PUT", "/samples/" + bson.NewObjectID().Hex(), sampleBody(missing, 7), 404},
		{"DELETE", "/samples/invalid", "", 400},
		{"GET", "/samples?station_id=invalid", "", 400},
		{"GET", "/alerts?station_id=invalid", "", 400},
		{"POST", "/alerts", "{}", 405},
		{"GET", "/stations", "", 200},
		{"POST", "/stations", `{"name":"N","river":"R","code":"C","location":{"city":"X","state":"PR"},"thresholds":[{"parameter":"pH","max":15}]}`, 400},
		{"GET", "/stations/invalid", "", 400},
		{"PUT", "/stations/invalid", "{}", 400},
		{"GET", "/unknown", "", 404},
	} {
		request, err := http.NewRequest(tc.method, base+tc.path, strings.NewReader(tc.body))
		require.NoError(t, err)
		request.Header.Set("Content-Type", "application/json")
		response, err := httpClient.Do(request)
		require.NoError(t, err)
		_, readErr := io.Copy(io.Discard, response.Body)
		closeErr := response.Body.Close()
		require.NoError(t, readErr)
		require.NoError(t, closeErr)
		require.Equal(t, tc.status, response.StatusCode, tc.method+" "+tc.path)
	}

	request := func(method, path, body string) (int, []byte, error) {
		req, err := http.NewRequest(method, base+path, strings.NewReader(body))
		if err != nil {
			return 0, nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := httpClient.Do(req)
		if err != nil {
			return 0, nil, err
		}
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		return resp.StatusCode, data, err
	}
	call := func(method, path, body string, want int, output any) {
		t.Helper()
		status, data, err := request(method, path, body)
		require.NoError(t, err)
		require.Equal(t, want, status, "%s %s: %s", method, path, data)
		if output != nil {
			require.NoError(t, json.Unmarshal(data, output))
		}
	}
	createStation := func(code string) string {
		t.Helper()
		body := fmt.Sprintf(`{"name":"Ponte","river":"Rio Ivaí","code":%q,"location":{"city":"Maringá","state":"PR"},"thresholds":[{"parameter":"pH","min":6,"max":9}]}`, code)
		var s station.Station
		call("POST", "/stations", body, 201, &s)
		return s.ID.Hex()
	}

	stationID := createStation("IVA-001")
	stationPath := "/stations/" + stationID
	alertsPath := "/alerts?station_id=" + stationID
	samplesPath := "/samples?station_id=" + stationID
	var valid, invalid sample.Sample
	call("POST", "/samples", sampleBody(stationID, 7.2), 201, &valid)
	require.Equal(t, "Rio Ivaí", valid.River)
	var alerts []alert.Alert
	call("GET", alertsPath, "", 200, &alerts)
	require.Equal(t, []alert.Alert{}, alerts)
	call("POST", "/samples", sampleBody(stationID, 4.2), 201, &invalid)
	call("GET", alertsPath, "", 200, &alerts)
	require.Len(t, alerts, 1)
	firstAlert := alerts[0]
	require.Equal(t, invalid.ID, firstAlert.SampleID)
	require.Contains(t, firstAlert.Message, "IVA-001")
	call("GET", "/samples?station_id=%ZZ", "", 400, nil)
	call("GET", "/alerts?station_id=%ZZ", "", 400, nil)

	var samples []sample.Sample
	call("GET", samplesPath, "", 200, &samples)
	require.Len(t, samples, 2)
	call("GET", "/samples?station_id="+missing, "", 200, &samples)
	require.Equal(t, []sample.Sample{}, samples)
	call("GET", "/alerts?station_id="+missing, "", 200, &alerts)
	require.Equal(t, []alert.Alert{}, alerts)
	call("DELETE", stationPath, "", 409, nil)

	invalidPath := "/samples/" + invalid.ID.Hex()
	call("PUT", invalidPath, sampleBody(stationID, 9.2), 200, nil)
	call("GET", "/alerts", "", 200, &alerts)
	require.Len(t, alerts, 1)
	require.Equal(t, firstAlert.ID, alerts[0].ID)
	require.Equal(t, firstAlert.CreatedAt, alerts[0].CreatedAt)
	require.Equal(t, 9.2, alerts[0].Value)
	call("PUT", invalidPath, sampleBody(stationID, 9), 200, nil)
	call("GET", alertsPath, "", 200, &alerts)
	require.Empty(t, alerts)

	db := client.Database(dbName)
	validator := func(rule bson.M) {
		t.Helper()
		err := db.RunCommand(ctx, bson.D{{Key: "collMod", Value: "alerts"}, {Key: "validator", Value: rule}}).Err()
		require.NoError(t, err)
	}
	validator(bson.M{"value": bson.M{"$lt": 0}})
	call("POST", "/samples", sampleBody(stationID, 4.2), 500, nil)
	call("PUT", invalidPath, sampleBody(stationID, 4.2), 500, nil)
	call("GET", samplesPath, "", 200, &samples)
	require.Len(t, samples, 2)
	for _, saved := range samples {
		if saved.ID == invalid.ID {
			require.Equal(t, 9.0, saved.Value)
		}
	}
	validator(bson.M{})

	call("PUT", invalidPath, sampleBody(stationID, 4.2), 200, nil)
	call("GET", alertsPath, "", 200, &alerts)
	require.Len(t, alerts, 1)
	require.NotEqual(t, firstAlert.ID, alerts[0].ID)
	call("DELETE", invalidPath, "", 204, nil)
	call("GET", alertsPath, "", 200, &alerts)
	require.Empty(t, alerts)
	call("DELETE", "/samples/"+valid.ID.Hex(), "", 204, nil)
	call("DELETE", stationPath, "", 204, nil)

	raceStation := createStation("RACE-001")
	raceCtx, cancelRace := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelRace()
	counted := make(chan struct{})
	resume := make(chan struct{})
	repo := sample.NewRepository(db)
	paused := false
	counter := sampleCounterFunc(func(ctx context.Context, id bson.ObjectID) (int64, error) {
		count, err := repo.CountByStation(ctx, id)
		if err != nil || paused {
			return count, err
		}
		// Pause the first delete after its snapshot sees an empty station.
		paused = true
		close(counted)
		select {
		case <-resume:
			return count, nil
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	})
	stationService := station.NewService(station.NewRepository(db), counter, database.MongoTransaction(client))
	deleted := make(chan error, 1)
	go func() {
		deleted <- stationService.DeleteStation(raceCtx, raceStation)
	}()
	select {
	case <-counted:
	case err := <-deleted:
		t.Fatalf("delete finished before counting samples: %v", err)
	case <-raceCtx.Done():
		t.Fatal(raceCtx.Err())
	}

	var concurrentSample sample.Sample
	call("POST", "/samples", sampleBody(raceStation, 4.2), 201, &concurrentSample)
	close(resume)
	select {
	case err := <-deleted:
		require.ErrorIs(t, err, station.ErrStationHasSamples)
	case <-raceCtx.Done():
		t.Fatal(raceCtx.Err())
	}
	call("GET", "/stations/"+raceStation, "", 200, nil)
	call("GET", "/samples?station_id="+raceStation, "", 200, &samples)
	require.Len(t, samples, 1)
	require.Equal(t, concurrentSample.ID, samples[0].ID)
	call("GET", "/alerts?station_id="+raceStation, "", 200, &alerts)
	require.Len(t, alerts, 1)
	require.Equal(t, concurrentSample.ID, alerts[0].SampleID)
}
