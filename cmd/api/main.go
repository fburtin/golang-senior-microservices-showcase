package main

import (
	"context"
	"net/http"
	"os"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/app"
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
		appLogger.Error("failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}

	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		appLogger.Error("failed to ping MongoDB", "error", err)
		os.Exit(1)
	}

	database := mongoClient.Database(cfg.MongoDatabase)

	customerRepository := repositories.NewMongoCustomerRepository(database)
	customerService := services.NewCustomerService(customerRepository)
	customerHandler := handlers.NewCustomerHandler(customerService)

	router := app.NewRouter(customerHandler, appLogger)

	appLogger.Info("HTTP server started", "port", cfg.Port)

	err = http.ListenAndServe(":"+cfg.Port, router)
	if err != nil {
		appLogger.Error("HTTP server stopped", "error", err)
		os.Exit(1)
	}
}
