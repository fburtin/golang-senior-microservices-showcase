package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/config"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/messaging"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/repositories"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, nil),
	)

	cfg := config.Load()

	logger.Info(
		"idempotent consumer configuration loaded",
		"mongo_database", cfg.MongoDatabase,
		"kafka_broker", cfg.KafkaBroker,
		"kafka_topic", cfg.KafkaCustomerTopic,
		"kafka_group_id", cfg.KafkaGroupID,
	)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	connectCtx, cancel := context.WithTimeout(
		ctx,
		cfg.MongoTimeout,
	)
	defer cancel()

	client, err := mongo.Connect(
		options.Client().ApplyURI(cfg.MongoURI),
	)
	if err != nil {
		logger.Error(
			"failed to connect to MongoDB",
			"error", err,
		)
		os.Exit(1)
	}

	defer func() {
		disconnectCtx, disconnectCancel := context.WithTimeout(
			context.Background(),
			cfg.MongoTimeout,
		)
		defer disconnectCancel()

		if err := client.Disconnect(disconnectCtx); err != nil {
			logger.Error(
				"failed to disconnect from MongoDB",
				"error", err,
			)
		}
	}()

	if err := client.Ping(connectCtx, nil); err != nil {
		logger.Error(
			"failed to ping MongoDB",
			"error", err,
		)
		os.Exit(1)
	}

	database := client.Database(cfg.MongoDatabase)

	inboxRepository :=
		repositories.NewMongoProcessedEventRepository(database)

	if err := inboxRepository.CreateIndexes(connectCtx); err != nil {
		logger.Error(
			"failed to create processed-event indexes",
			"error", err,
		)
		os.Exit(1)
	}

	handler :=
		messaging.NewLoggingCustomerCreatedHandler(logger)

	consumer := messaging.NewCustomerConsumer(
		cfg.KafkaBroker,
		cfg.KafkaGroupID,
		cfg.KafkaCustomerTopic,
		inboxRepository,
		handler,
		logger,
	)

	defer func() {
		if err := consumer.Close(); err != nil {
			logger.Error(
				"failed to close Kafka consumer",
				"error", err,
			)
		}
	}()

	if err := consumer.Start(ctx); err != nil && ctx.Err() == nil {
		logger.Error(
			"Kafka consumer stopped unexpectedly",
			"error", err,
		)
		os.Exit(1)
	}

	logger.Info("idempotent consumer stopped")
}
