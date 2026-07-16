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

func (r *MongoOutboxRepository) ClaimPending(
	ctx context.Context,
	workerID string,
	limit int64,
) ([]domain.OutboxEvent, error) {
	now := time.Now().UTC()
	events := make([]domain.OutboxEvent, 0, limit)

	for int64(len(events)) < limit {
		filter := bson.D{
			{Key: "status", Value: domain.OutboxPending},
			{
				Key: "$or",
				Value: bson.A{
					bson.D{
						{
							Key: "nextAttemptAt",
							Value: bson.D{
								{Key: "$exists", Value: false},
							},
						},
					},
					bson.D{
						{
							Key: "nextAttemptAt",
							Value: bson.D{
								{Key: "$lte", Value: now},
							},
						},
					},
				},
			},
		}

		update := bson.D{
			{
				Key: "$set",
				Value: bson.D{
					{Key: "status", Value: domain.OutboxProcessing},
					{Key: "lockedAt", Value: now},
					{Key: "lockedBy", Value: workerID},
				},
			},
		}

		findOptions := options.FindOneAndUpdate().
			SetSort(bson.D{{Key: "createdAt", Value: 1}}).
			SetReturnDocument(options.After)

		var event domain.OutboxEvent

		err := r.collection.
			FindOneAndUpdate(ctx, filter, update, findOptions).
			Decode(&event)

		if errors.Is(err, mongo.ErrNoDocuments) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf(
				"claim pending outbox event: %w",
				err,
			)
		}

		events = append(events, event)
	}

	return events, nil
}

func (r *MongoOutboxRepository) MarkPublished(
	ctx context.Context,
	id string,
) error {
	now := time.Now().UTC()

	filter := bson.D{
		{Key: "id", Value: id},
		{Key: "status", Value: domain.OutboxProcessing},
	}

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
				{Key: "lockedAt", Value: ""},
				{Key: "lockedBy", Value: ""},
				{Key: "nextAttemptAt", Value: ""},
			},
		},
	}

	return r.updateOne(ctx, filter, update)
}

func (r *MongoOutboxRepository) RecordFailure(
	ctx context.Context,
	id string,
	reason string,
	nextAttemptAt time.Time,
) error {
	filter := bson.D{
		{Key: "id", Value: id},
		{Key: "status", Value: domain.OutboxProcessing},
	}

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
				{Key: "status", Value: domain.OutboxPending},
				{Key: "lastError", Value: reason},
				{Key: "nextAttemptAt", Value: nextAttemptAt},
			},
		},
		{
			Key: "$unset",
			Value: bson.D{
				{Key: "lockedAt", Value: ""},
				{Key: "lockedBy", Value: ""},
			},
		},
	}

	return r.updateOne(ctx, filter, update)
}

func (r *MongoOutboxRepository) MarkFailed(
	ctx context.Context,
	id string,
	reason string,
) error {
	filter := bson.D{
		{Key: "id", Value: id},
	}

	update := bson.D{
		{
			Key: "$set",
			Value: bson.D{
				{Key: "status", Value: domain.OutboxFailed},
				{Key: "lastError", Value: reason},
			},
		},
		{
			Key: "$unset",
			Value: bson.D{
				{Key: "lockedAt", Value: ""},
				{Key: "lockedBy", Value: ""},
				{Key: "nextAttemptAt", Value: ""},
			},
		},
	}

	return r.updateOne(ctx, filter, update)
}

func (r *MongoOutboxRepository) ReleaseStaleLocks(
	ctx context.Context,
	staleBefore time.Time,
) error {
	filter := bson.D{
		{Key: "status", Value: domain.OutboxProcessing},
		{
			Key: "lockedAt",
			Value: bson.D{
				{Key: "$lte", Value: staleBefore},
			},
		},
	}

	update := bson.D{
		{
			Key: "$set",
			Value: bson.D{
				{Key: "status", Value: domain.OutboxPending},
			},
		},
		{
			Key: "$unset",
			Value: bson.D{
				{Key: "lockedAt", Value: ""},
				{Key: "lockedBy", Value: ""},
			},
		},
	}

	_, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("release stale outbox locks: %w", err)
	}

	return nil
}

func (r *MongoOutboxRepository) CreateIndexes(
	ctx context.Context,
) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "nextAttemptAt", Value: 1},
				{Key: "createdAt", Value: 1},
			},
			Options: options.Index().
				SetName("idx_outbox_status_next_attempt_created_at"),
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "lockedAt", Value: 1},
			},
			Options: options.Index().
				SetName("idx_outbox_status_locked_at"),
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

func (r *MongoOutboxRepository) updateOne(
	ctx context.Context,
	filter bson.D,
	update bson.D,
) error {
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("update outbox event: %w", err)
	}

	if result.MatchedCount == 0 {
		return ErrOutboxEventNotFound
	}

	return nil
}
