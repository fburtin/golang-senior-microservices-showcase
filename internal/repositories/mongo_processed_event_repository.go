package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const processedEventsCollectionName = "processed_events"

type MongoProcessedEventRepository struct {
	collection *mongo.Collection
}

func NewMongoProcessedEventRepository(database *mongo.Database) *MongoProcessedEventRepository {
	return &MongoProcessedEventRepository{
		collection: database.Collection(processedEventsCollectionName),
	}
}

func (r *MongoProcessedEventRepository) TryStart(
	ctx context.Context,
	eventID string,
	eventType string,
) (bool, error) {
	record := domain.ProcessedEvent{
		EventID:   eventID,
		EventType: eventType,
		Status:    domain.ProcessedEventProcessing,
		StartedAt: time.Now().UTC(),
	}

	_, err := r.collection.InsertOne(ctx, record)
	if err == nil {
		return true, nil
	}

	if mongo.IsDuplicateKeyError(err) {
		return false, nil
	}

	return false, fmt.Errorf("reserve processed event %q: %w", eventID, err)
}

func (r *MongoProcessedEventRepository) MarkCompleted(
	ctx context.Context,
	eventID string,
) error {
	now := time.Now().UTC()

	result, err := r.collection.UpdateOne(
		ctx,
		bson.D{{Key: "eventId", Value: eventID}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "status", Value: domain.ProcessedEventCompleted},
			{Key: "completedAt", Value: now},
		}}},
	)
	if err != nil {
		return fmt.Errorf("complete processed event %q: %w", eventID, err)
	}

	if result.MatchedCount == 0 {
		return errors.New("processed event not found")
	}

	return nil
}

func (r *MongoProcessedEventRepository) Remove(
	ctx context.Context,
	eventID string,
) error {
	_, err := r.collection.DeleteOne(
		ctx,
		bson.D{{Key: "eventId", Value: eventID}},
	)
	if err != nil {
		return fmt.Errorf("remove processed event %q: %w", eventID, err)
	}

	return nil
}

func (r *MongoProcessedEventRepository) CreateIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.D{{Key: "eventId", Value: 1}},
			Options: options.Index().
				SetName("ux_processed_events_event_id").
				SetUnique(true),
		},
	)
	if err != nil {
		return fmt.Errorf("create processed-event indexes: %w", err)
	}

	return nil
}
