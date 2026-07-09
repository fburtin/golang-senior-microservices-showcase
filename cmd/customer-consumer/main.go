package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/config"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/messaging"
)

func main() {

	// Load configuration
	cfg := config.Load()

	log.Printf("Broker: %s", cfg.KafkaBroker)
	log.Printf("Topic : '%s'", cfg.KafkaCustomerTopic)

	// Create Kafka consumer
	consumer := messaging.NewCustomerConsumer(
		cfg.KafkaBroker,
		"customer-consumer-group",
		cfg.KafkaCustomerTopic,
	)

	// Handle Ctrl+C
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	log.Println("===================================")
	log.Println(" Customer Consumer Started")
	log.Println("===================================")
	log.Printf("Broker : %s", cfg.KafkaBroker)
	log.Printf("Topic  : %s", cfg.KafkaCustomerTopic)
	log.Println()

	// Start consuming messages
	if err := consumer.Start(ctx); err != nil {
		log.Fatal(err)
	}
}
