package repositories

import (
	"context"
	"time"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/domain"
)

type OutboxRepository interface {
	ClaimPending(
		ctx context.Context,
		workerID string,
		limit int64,
	) ([]domain.OutboxEvent, error)

	MarkPublished(
		ctx context.Context,
		id string,
	) error

	RecordFailure(
		ctx context.Context,
		id string,
		reason string,
		nextAttemptAt time.Time,
	) error

	MarkFailed(
		ctx context.Context,
		id string,
		reason string,
	) error

	ReleaseStaleLocks(
		ctx context.Context,
		staleBefore time.Time,
	) error

	CreateIndexes(ctx context.Context) error
}
