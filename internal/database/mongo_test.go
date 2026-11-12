package database

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBuildMongoURI(t *testing.T) {
	for _, tc := range []struct{ name, uri, user, password, port, want string }{
		{"defaults", "", "", "", "", "mongodb://localhost:27017/?directConnection=true&replicaSet=rs0"},
		{"custom port", "", "", "", "27018", "mongodb://localhost:27018/?directConnection=true&replicaSet=rs0"},
		{"credentials", "", "testuser", "testpass", "", "mongodb://testuser:testpass@localhost:27017/?authSource=admin&directConnection=true&replicaSet=rs0"},
		{"credentials and port", "", "testuser", "testpass", "27018", "mongodb://testuser:testpass@localhost:27018/?authSource=admin&directConnection=true&replicaSet=rs0"},
		{"user only", "", "testuser", "", "", "mongodb://localhost:27017/?directConnection=true&replicaSet=rs0"},
		{"password only", "", "", "testpass", "", "mongodb://localhost:27017/?directConnection=true&replicaSet=rs0"},
		{"explicit URI wins", "mongodb://test-host:27019", "ignored", "ignored", "27018", "mongodb://test-host:27019"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("MONGO_URI", tc.uri)
			t.Setenv("MONGO_INITDB_ROOT_USERNAME", tc.user)
			t.Setenv("MONGO_INITDB_ROOT_PASSWORD", tc.password)
			t.Setenv("MONGO_PORT", tc.port)
			t.Setenv("MONGO_HOST", "")
			require.Equal(t, tc.want, BuildMongoURI())
		})
	}
}

func TestConnectMongoDBInvalidURI(t *testing.T) {
	client, err := ConnectMongoDB("invalid://host")
	require.Nil(t, client)
	require.ErrorContains(t, err, "failed to instantiate MongoDB client")
}

func TestConnectMongoDBIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("MongoDB integration: run without -short for full coverage")
	}
	uri := os.Getenv("MONGO_TEST_URI")
	require.NotEmpty(t, uri, "set MONGO_TEST_URI and start the test MongoDB")
	client, err := ConnectMongoDB(uri)
	require.NoError(t, err)
	require.NotNil(t, client)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		require.NoError(t, client.Disconnect(ctx))
	})
}
