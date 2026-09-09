package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

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
	httpClient := &http.Client{Timeout: time.Second}
	defer httpClient.CloseIdleConnections()
	require.Eventually(t, func() bool {
		response, err := httpClient.Get(base + "/samples")
		if err != nil {
			return false
		}
		defer response.Body.Close()
		return response.StatusCode == http.StatusOK
	}, 10*time.Second, 50*time.Millisecond, "API did not become ready")
	for _, tc := range []struct {
		method, path, body string
		status             int
	}{
		{"GET", "/swagger/index.html", "", 200},
		{"POST", "/samples", `{"river":"Rio","parameter":"pH","value":15}`, 400},
		{"PUT", "/samples/invalid", `{"river":"Rio","parameter":"pH","value":7}`, 400},
		{"DELETE", "/samples/invalid", "", 400},
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
}
