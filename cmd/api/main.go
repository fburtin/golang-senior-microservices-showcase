package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/config"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/handlers"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/logger"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/repositories"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/services"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {

	cfg := config.Load()
	appLogger := logger.New()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.MongoTimeout)
	defer cancel()
	mongoClient, err := mongo.Connect(options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		appLogger.Error(
			"failed to connect to MongoDB",
			"error", err,
		)
		os.Exit(1)
	}
	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		appLogger.Error(
			"failed to ping MongoDB",
			"error", err,
		)
		os.Exit(1)
	}

	database := mongoClient.Database(cfg.MongoDatabase)

	customerRepository := repositories.NewMongoCustomerRepository(database)
	customerService := services.NewCustomerService(customerRepository)
	customerHandler := handlers.NewCustomerHandler(customerService)

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/customers", customerHandler.HandleCustomers)
	http.HandleFunc("/customers/{id}", customerHandler.HandleCustomerByID)

	appLogger.Info(
		"HTTP server started",
		"port", cfg.Port,
	)

	err = http.ListenAndServe(":"+cfg.Port, nil)
	if err != nil {
		appLogger.Error(
			"HTTP server stopped",
			"error", err,
		)
		os.Exit(1)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "OK",
		"service": "Go Senior Microservices Showcase",
		"version": "1.0.0",
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(value)
}
