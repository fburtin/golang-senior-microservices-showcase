package app

import (
	"context"
	"log/slog"
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

type App struct {
	config config.Config
	logger *slog.Logger
	router http.Handler
}

func New() *App {
	cfg := config.Load()
	appLogger := logger.New()

	database := connectMongo(cfg, appLogger)

	customerRepository := repositories.NewMongoCustomerRepository(database)
	customerService := services.NewCustomerService(customerRepository)
	customerHandler := handlers.NewCustomerHandler(customerService)

	router := NewRouter(customerHandler, appLogger)

	return &App{
		config: cfg,
		logger: appLogger,
		router: router,
	}
}

func (a *App) Run() {
	a.logger.Info("HTTP server started", "port", a.config.Port)

	err := http.ListenAndServe(":"+a.config.Port, a.router)
	if err != nil {
		a.logger.Error("HTTP server stopped", "error", err)
		os.Exit(1)
	}
}

func connectMongo(cfg config.Config, appLogger *slog.Logger) *mongo.Database {
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

	return mongoClient.Database(cfg.MongoDatabase)
}
