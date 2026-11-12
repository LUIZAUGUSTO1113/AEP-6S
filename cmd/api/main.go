package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/LUIZAUGUSTO1113/AEP-6S/internal/alert"
	"github.com/LUIZAUGUSTO1113/AEP-6S/internal/database"
	"github.com/LUIZAUGUSTO1113/AEP-6S/internal/sample"
	"github.com/LUIZAUGUSTO1113/AEP-6S/internal/station"
	"github.com/joho/godotenv"

	_ "github.com/LUIZAUGUSTO1113/AEP-6S/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title Water Quality Monitoring API
// @version 1.0
// @description This is a sample server for monitoring water quality in rivers.
// @BasePath /
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("[INFO] No .env file found, using system environment variables")
	}

	fmt.Println("Starting Water Quality Monitoring API...")

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

	db := client.Database(dbName)
	sampleRepository := sample.NewRepository(db)
	stationRepository := station.NewRepository(db)
	alertRepository := alert.NewRepository(db)
	indexContext, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	for _, ensure := range []func(context.Context) error{
		stationRepository.EnsureIndexes, sampleRepository.EnsureIndexes, alertRepository.EnsureIndexes,
	} {
		if err := ensure(indexContext); err != nil {
			cancel()
			log.Fatalf("[ERROR] Failed to initialize MongoDB indexes: %v", err)
		}
	}
	cancel()

	transaction := database.MongoTransaction(client)
	sampleController := sample.NewController(sample.NewService(sampleRepository, stationRepository, alertRepository, transaction))
	stationController := station.NewController(station.NewService(stationRepository, sampleRepository, transaction))
	alertController := alert.NewController(alert.NewService(alertRepository))

	mux := http.NewServeMux()

	mux.HandleFunc("POST /stations", stationController.Create)
	mux.HandleFunc("GET /stations", stationController.GetAll)
	mux.HandleFunc("GET /stations/{id}", stationController.GetByID)
	mux.HandleFunc("PUT /stations/{id}", stationController.Update)
	mux.HandleFunc("DELETE /stations/{id}", stationController.DeleteByID)

	mux.HandleFunc("POST /samples", sampleController.Create)
	mux.HandleFunc("GET /samples", sampleController.GetAll)
	mux.HandleFunc("PUT /samples/{id}", sampleController.Update)
	mux.HandleFunc("DELETE /samples/{id}", sampleController.DeleteById)

	mux.HandleFunc("GET /alerts", alertController.GetAll)

	mux.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("[INFO] Starting server on: http://localhost:%s\n", port)
	fmt.Printf("[INFO] Swagger documentation available at: http://localhost:%s/swagger/index.html\n", port)

	if err := http.ListenAndServe(":"+port, allowVite(mux)); err != nil {
		log.Fatalf("[ERROR] Failed to start server: %v", err)
	}
}
