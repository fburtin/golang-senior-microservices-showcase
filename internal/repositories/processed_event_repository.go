package repositories

import "context"

type ProcessedEventRepository interface {
	// TryStart atomically reserves an EventID.
	// It returns false, nil when the event already exists.
	TryStart(ctx context.Context, eventID, eventType string) (bool, error)
	MarkCompleted(ctx context.Context, eventID string) error
	Remove(ctx context.Context, eventID string) error
	CreateIndexes(ctx context.Context) error
}
