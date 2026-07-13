package workers

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/domain"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/messaging"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/repositories"
)

const (
	defaultBatchSize   = int64(50)
	maxPublishAttempts = 5
)

type CustomerEventProducer interface {
	PublishCustomerCreated(
		ctx context.Context,
		event messaging.CustomerCreatedEvent,
	) error
}

type OutboxWorker struct {
	repository repositories.OutboxRepository
	producer   CustomerEventProducer
	logger     *slog.Logger
	interval   time.Duration
}

func NewOutboxWorker(
	repository repositories.OutboxRepository,
	producer CustomerEventProducer,
	logger *slog.Logger,
	interval time.Duration,
) *OutboxWorker {
	return &OutboxWorker{
		repository: repository,
		producer:   producer,
		logger:     logger,
		interval:   interval,
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
	events, err := w.repository.FindPending(
		ctx,
		defaultBatchSize,
	)
	if err != nil {
		w.logger.Error(
			"failed to load pending outbox events",
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
	var event messaging.CustomerCreatedEvent

	if err := json.Unmarshal(outboxEvent.Payload, &event); err != nil {
		_ = w.repository.MarkFailed(
			ctx,
			outboxEvent.ID,
			err.Error(),
		)
		return
	}

	if err := w.producer.PublishCustomerCreated(ctx, event); err != nil {
		_ = w.repository.RecordFailure(
			ctx,
			outboxEvent.ID,
			err.Error(),
		)

		if outboxEvent.Attempts+1 >= maxPublishAttempts {
			_ = w.repository.MarkFailed(
				ctx,
				outboxEvent.ID,
				err.Error(),
			)
		}

		return
	}

	if err := w.repository.MarkPublished(
		ctx,
		outboxEvent.ID,
	); err != nil {
		w.logger.Error(
			"event published but not marked as published",
			"event_id",
			outboxEvent.EventID,
			"error",
			err,
		)
	}
}
