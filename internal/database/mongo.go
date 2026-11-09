package database

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func BuildMongoURI() string {
	if uri := os.Getenv("MONGO_URI"); uri != "" {
		return uri
	}

	host := os.Getenv("MONGO_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("MONGO_PORT")
	if port == "" {
		port = "27017"
	}

	uri := url.URL{Scheme: "mongodb", Host: net.JoinHostPort(host, port), Path: "/"}
	query := url.Values{"directConnection": {"true"}, "replicaSet": {"rs0"}}
	username := os.Getenv("MONGO_INITDB_ROOT_USERNAME")
	password := os.Getenv("MONGO_INITDB_ROOT_PASSWORD")
	if username != "" && password != "" {
		uri.User = url.UserPassword(username, password)
		query.Set("authSource", "admin")
	}
	uri.RawQuery = query.Encode()
	return uri.String()
}

func ConnectMongoDB(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate MongoDB client: %w", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("failed to authenticate/connect with MongoDB: %w", err)
	}

	return client, nil
}
