package database

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Transaction func(context.Context, func(context.Context) error) error

func MongoTransaction(client *mongo.Client) Transaction {
	return func(ctx context.Context, operation func(context.Context) error) error {
		session, err := client.StartSession()
		if err != nil {
			return err
		}
		defer session.EndSession(ctx)

		_, err = session.WithTransaction(ctx, func(ctx context.Context) (any, error) {
			return nil, operation(ctx)
		})
		return err
	}
}
