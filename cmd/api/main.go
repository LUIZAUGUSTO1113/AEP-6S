package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/LUIZAUGUSTO1113/AEP-6S/internal/database"
	"github.com/LUIZAUGUSTO1113/AEP-6S/internal/sample"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err := godotenv.Load(); err != nil {
		log.Println("[INFO] No .env file found, using system environment variables")
	}

	fmt.Println("====")
	fmt.Println("PoC: Monitoramento de Qualidade da Água em Rios")
	fmt.Println("====")

	uri := database.BuildMongoURI()
	client, err := database.ConnectMongoDB(uri)
	if err != nil {
		log.Fatalf("[ERROR] Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	fmt.Println("[INFO] Successfully connected to MongoDB")

	dbName := os.Getenv("MONGO_INITDB_DATABASE")
	if dbName == "" {
		dbName = "water_monitoring"
	}

	database := client.Database(dbName)
	repository := sample.NewRepository(database)
	ctx := context.Background()

	sampleData := &sample.Sample{
		River:       "Rio Paraná",
		Parameter:   "pH",
		Value:       7.2,
		CollectedAt: time.Now(),
	}

	if err := repository.Create(ctx, sampleData); err != nil {
		log.Fatalf("[ERROR] Failed to create sample: %v", err)
	}
	fmt.Printf("[INFO] Sample created successfully - ID: %s", sampleData.ID.Hex())

	samples, err := repository.FindAll(ctx)
	if err != nil {
		log.Fatalf("[ERROR] Failed to retrieve samples: %v", err)
	}
	fmt.Printf("[INFO] Retrieved %d samples\n", len(samples))
	for _, s := range samples {
		fmt.Printf("River: %s, Parameter: %s, Value: %.2f, CollectedAt: %s\n", s.River, s.Parameter, s.Value, s.CollectedAt.Format(time.RFC3339))
	}
}
