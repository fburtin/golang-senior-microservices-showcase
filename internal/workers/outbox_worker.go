package workers

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/domain"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/messaging"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/repositories"
)

const (
	defaultBatchSize   = int64(50)
	maxPublishAttempts = 5
	baseRetryDelay     = 5 * time.Second
	maxRetryDelay      = 5 * time.Minute
)

type CustomerEventProducer interface {
	PublishCustomerCreated(
		ctx context.Context,
		event messaging.CustomerCreatedEvent,
	) error
}

type OutboxWorker struct {
	repository  repositories.OutboxRepository
	producer    CustomerEventProducer
	logger      *slog.Logger
	interval    time.Duration
	workerID    string
	lockTimeout time.Duration

	// publishedEventIDs prevents an event from being sent twice by this worker
	// process when Kafka publishing succeeds but MarkPublished fails.
	//
	// The EventID is stable across retries. sync.Map is used because Run may
	// later be extended to process events concurrently.
	publishedEventIDs sync.Map
}

func NewOutboxWorker(
	repository repositories.OutboxRepository,
	producer CustomerEventProducer,
	logger *slog.Logger,
	interval time.Duration,
	workerID string,
	lockTimeout time.Duration,
) *OutboxWorker {
	return &OutboxWorker{
		repository:  repository,
		producer:    producer,
		logger:      logger,
		interval:    interval,
		workerID:    workerID,
		lockTimeout: lockTimeout,
	}
}

func (w *OutboxWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.processBatch(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("outbox worker stopped")
			return

		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *OutboxWorker) processBatch(ctx context.Context) {
	staleBefore := time.Now().UTC().Add(-w.lockTimeout)

	if err := w.repository.ReleaseStaleLocks(
		ctx,
		staleBefore,
	); err != nil {
		w.logger.Error(
			"failed to release stale outbox locks",
			"error",
			err,
		)
	}

	events, err := w.repository.ClaimPending(
		ctx,
		w.workerID,
		defaultBatchSize,
	)
	if err != nil {
		w.logger.Error(
			"failed to claim pending outbox events",
			"error",
			err,
		)
		return
	}

	for _, event := range events {
		w.processEvent(ctx, event)
	}
}

func (w *OutboxWorker) processEvent(
	ctx context.Context,
	outboxEvent domain.OutboxEvent,
) {
	eventID := outboxEvent.EventID
	if eventID == "" {
		// Defensive fallback for legacy outbox records.
		eventID = outboxEvent.ID
	}

	// Kafka may already have accepted this event during an earlier attempt while
	// MongoDB failed to persist the PUBLISHED state. Do not publish it again
	// within the same worker process; retry only the state transition.
	if w.wasPublished(eventID) {
		w.markPublished(ctx, outboxEvent)
		return
	}

	var event messaging.CustomerCreatedEvent

	if err := json.Unmarshal(outboxEvent.Payload, &event); err != nil {
		w.markFailed(ctx, outboxEvent, err.Error())
		return
	}

	// Keep the payload EventID aligned with the durable outbox identifier so
	// consumers can use it as their idempotency key.
	event.EventID = eventID

	if err := w.producer.PublishCustomerCreated(ctx, event); err != nil {
		nextAttempt := outboxEvent.Attempts + 1

		if nextAttempt >= maxPublishAttempts {
			w.markFailed(ctx, outboxEvent, err.Error())
			return
		}

		nextAttemptAt := time.Now().UTC().Add(
			retryDelay(outboxEvent.Attempts),
		)

		if recordErr := w.repository.RecordFailure(
			ctx,
			outboxEvent.ID,
			err.Error(),
			nextAttemptAt,
		); recordErr != nil {
			w.logger.Error(
				"failed to record outbox publishing failure",
				"event_id",
				eventID,
				"error",
				recordErr,
			)
		}

		return
	}

	// Store the idempotency key before updating MongoDB. If MarkPublished fails,
	// a stale-lock retry in this process will skip Kafka and retry MongoDB only.
	w.publishedEventIDs.Store(eventID, struct{}{})
	w.markPublished(ctx, outboxEvent)
}

func (w *OutboxWorker) markPublished(
	ctx context.Context,
	event domain.OutboxEvent,
) {
	if err := w.repository.MarkPublished(
		ctx,
		event.ID,
	); err != nil {
		w.logger.Error(
			"event published but not marked as published",
			"event_id",
			event.EventID,
			"error",
			err,
		)
		return
	}

	eventID := event.EventID
	if eventID == "" {
		eventID = event.ID
	}
	w.publishedEventIDs.Delete(eventID)
}

func (w *OutboxWorker) wasPublished(eventID string) bool {
	_, ok := w.publishedEventIDs.Load(eventID)
	return ok
}

func (w *OutboxWorker) markFailed(
	ctx context.Context,
	event domain.OutboxEvent,
	reason string,
) {
	if err := w.repository.MarkFailed(
		ctx,
		event.ID,
		reason,
	); err != nil {
		w.logger.Error(
			"failed to mark outbox event as failed",
			"event_id",
			event.EventID,
			"error",
			err,
		)
	}
}

func retryDelay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	delay := baseRetryDelay * time.Duration(1<<attempt)
	if delay > maxRetryDelay {
		return maxRetryDelay
	}

	return delay
}
