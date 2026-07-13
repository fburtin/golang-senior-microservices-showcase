package repositories

import (
	"context"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/domain"
)

type OutboxRepository interface {
	FindPending(
		ctx context.Context,
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
	) error

	MarkFailed(
		ctx context.Context,
		id string,
		reason string,
	) error

	CreateIndexes(ctx context.Context) error
}
