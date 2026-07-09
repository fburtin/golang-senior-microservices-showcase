package messaging

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type CustomerConsumer struct {
	reader *kafka.Reader
}

func NewCustomerConsumer(broker string, groupID string, topic string) *CustomerConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{broker},
		Topic:       topic,
		GroupID:     groupID,
		StartOffset: kafka.FirstOffset,
		MinBytes:    1,
		MaxBytes:    10e6,
		Logger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			log.Printf(msg, args...)
		}),
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			log.Printf(msg, args...)
		}),
	})

	return &CustomerConsumer{
		reader: reader,
	}
}

func (c *CustomerConsumer) Start(ctx context.Context) error {
	log.Println("Kafka consumer started...")

	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("Kafka consumer stopped")
				return c.reader.Close()
			}

			log.Printf("Kafka consumer error: %v", err)
			continue
		}

		log.Printf(
			"Received message topic=%s partition=%d offset=%d key=%s value=%s",
			msg.Topic,
			msg.Partition,
			msg.Offset,
			string(msg.Key),
			string(msg.Value),
		)
	}
}
