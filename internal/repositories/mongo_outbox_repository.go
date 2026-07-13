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

const outboxCollectionName = "outbox_events"

var ErrOutboxEventNotFound = errors.New("outbox event not found")

type MongoOutboxRepository struct {
	collection *mongo.Collection
}

func NewMongoOutboxRepository(
	database *mongo.Database,
) *MongoOutboxRepository {
	return &MongoOutboxRepository{
		collection: database.Collection(outboxCollectionName),
	}
}

func (r *MongoOutboxRepository) FindPending(
	ctx context.Context,
	limit int64,
) ([]domain.OutboxEvent, error) {
	filter := bson.D{
		{Key: "status", Value: domain.OutboxPending},
	}

	findOptions := options.Find().
		SetLimit(limit).
		SetSort(bson.D{
			{Key: "createdAt", Value: 1},
		})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("find pending outbox events: %w", err)
	}
	defer cursor.Close(ctx)

	events := make([]domain.OutboxEvent, 0)

	if err := cursor.All(ctx, &events); err != nil {
		return nil, fmt.Errorf("decode outbox events: %w", err)
	}

	return events, nil
}

func (r *MongoOutboxRepository) MarkPublished(
	ctx context.Context,
	id string,
) error {
	now := time.Now().UTC()

	update := bson.D{
		{
			Key: "$set",
			Value: bson.D{
				{Key: "status", Value: domain.OutboxPublished},
				{Key: "publishedAt", Value: now},
			},
		},
		{
			Key: "$unset",
			Value: bson.D{
				{Key: "lastError", Value: ""},
			},
		},
	}

	return r.updateByID(ctx, id, update)
}

func (r *MongoOutboxRepository) RecordFailure(
	ctx context.Context,
	id string,
	reason string,
) error {
	update := bson.D{
		{
			Key: "$inc",
			Value: bson.D{
				{Key: "attempts", Value: 1},
			},
		},
		{
			Key: "$set",
			Value: bson.D{
				{Key: "lastError", Value: reason},
			},
		},
	}

	return r.updateByID(ctx, id, update)
}

func (r *MongoOutboxRepository) MarkFailed(
	ctx context.Context,
	id string,
	reason string,
) error {
	update := bson.D{
		{
			Key: "$set",
			Value: bson.D{
				{Key: "status", Value: domain.OutboxFailed},
				{Key: "lastError", Value: reason},
			},
		},
	}

	return r.updateByID(ctx, id, update)
}

func (r *MongoOutboxRepository) CreateIndexes(
	ctx context.Context,
) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "createdAt", Value: 1},
			},
			Options: options.Index().
				SetName("idx_outbox_status_created_at"),
		},
		{
			Keys: bson.D{
				{Key: "eventId", Value: 1},
			},
			Options: options.Index().
				SetName("ux_outbox_event_id").
				SetUnique(true),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("create outbox indexes: %w", err)
	}

	return nil
}

func (r *MongoOutboxRepository) updateByID(
	ctx context.Context,
	id string,
	update bson.D,
) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.D{{Key: "id", Value: id}},
		update,
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrOutboxEventNotFound
	}

	return nil
}
