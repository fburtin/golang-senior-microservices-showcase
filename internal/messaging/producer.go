package messaging

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(broker, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(broker),
			Topic:    topic,
			Balancer: &kafka.Hash{},
		},
	}
}

func (p *Producer) PublishCustomerCreated(
	ctx context.Context,
	event CustomerCreatedEvent,
) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// The durable EventID is reused on every retry. Kafka remains at-least-once,
	// while consumers use this key for durable deduplication.
	return p.writer.WriteMessages(
		ctx,
		kafka.Message{
			Key:   []byte(event.EventID),
			Value: data,
			Headers: []kafka.Header{
				{Key: "event-id", Value: []byte(event.EventID)},
				{Key: "event-type", Value: []byte(event.EventType)},
			},
		},
	)
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
