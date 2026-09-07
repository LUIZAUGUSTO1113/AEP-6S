package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/LUIZAUGUSTO1113/AEP-6S/internal/database"
	"github.com/LUIZAUGUSTO1113/AEP-6S/internal/sample"
	"github.com/joho/godotenv"

	_ "github.com/LUIZAUGUSTO1113/AEP-6S/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title Water Quality Monitoring API
// @version 1.0
// @description This is a sample server for monitoring water quality in rivers.
// @BasePath /
func main() {
	err := godotenv.Load()
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

	database := client.Database(dbName)
	repository := sample.NewRepository(database)
	service := sample.NewService(repository)
	controller := sample.NewController(service)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /samples", controller.Create)
	mux.HandleFunc("GET /samples", controller.GetAll)
	mux.HandleFunc("DELETE /samples/{id}", controller.DeleteById)

	mux.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("[INFO] Starting server on: http://localhost:%s\n", port)
	fmt.Printf("[INFO] Swagger documentation available at: http://localhost:%s/swagger/index.html\n", port)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("[ERROR] Failed to start server: %v", err)
	}
}
