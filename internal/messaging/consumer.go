package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/repositories"
	"github.com/segmentio/kafka-go"
)

type CustomerCreatedHandler interface {
	HandleCustomerCreated(ctx context.Context, event CustomerCreatedEvent) error
}

type CustomerConsumer struct {
	reader  *kafka.Reader
	inbox   repositories.ProcessedEventRepository
	handler CustomerCreatedHandler
	logger  *slog.Logger
}

func NewCustomerConsumer(
	broker string,
	groupID string,
	topic string,
	inbox repositories.ProcessedEventRepository,
	handler CustomerCreatedHandler,
	logger *slog.Logger,
) *CustomerConsumer {
	return &CustomerConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     []string{broker},
			Topic:       topic,
			GroupID:     groupID,
			StartOffset: kafka.FirstOffset,
			MinBytes:    1,
			MaxBytes:    10e6,
		}),
		inbox:   inbox,
		handler: handler,
		logger:  logger,
	}
}

func (c *CustomerConsumer) Start(ctx context.Context) error {
	c.logger.Info("Kafka consumer started")

	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return c.reader.Close()
			}
			c.logger.Error("fetch Kafka message", "error", err)
			continue
		}

		if err := c.processMessage(ctx, message); err != nil {
			// Do not commit: Kafka will redeliver the same EventID.
			c.logger.Error(
				"process Kafka message",
				"partition", message.Partition,
				"offset", message.Offset,
				"error", err,
			)
			continue
		}

		if err := c.reader.CommitMessages(ctx, message); err != nil {
			// A failed commit may produce a duplicate delivery. The inbox makes
			// that redelivery safe.
			c.logger.Error("commit Kafka message", "error", err)
		}
	}
}

func (c *CustomerConsumer) processMessage(
	ctx context.Context,
	message kafka.Message,
) error {
	var event CustomerCreatedEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		return fmt.Errorf("decode customer-created event: %w", err)
	}

	if event.EventID == "" {
		event.EventID = string(message.Key)
	}
	if event.EventID == "" {
		return fmt.Errorf("eventId is required")
	}

	reserved, err := c.inbox.TryStart(ctx, event.EventID, event.EventType)
	if err != nil {
		return err
	}

	if !reserved {
		c.logger.Info("duplicate event ignored", "event_id", event.EventID)
		return nil
	}

	if err := c.handler.HandleCustomerCreated(ctx, event); err != nil {
		// Release the reservation so the same EventID can be retried. The
		// business handler must execute its own database side effects in the
		// same transaction as the inbox record for strict atomicity.
		if removeErr := c.inbox.Remove(ctx, event.EventID); removeErr != nil {
			return fmt.Errorf(
				"handle event: %v; remove inbox reservation: %w",
				err,
				removeErr,
			)
		}
		return fmt.Errorf("handle customer-created event: %w", err)
	}

	if err := c.inbox.MarkCompleted(ctx, event.EventID); err != nil {
		return err
	}

	return nil
}

func (c *CustomerConsumer) Close() error {
	return c.reader.Close()
}
