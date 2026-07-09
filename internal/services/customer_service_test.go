package services

import (
	"context"
	"errors"
	"testing"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/domain"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/messaging"
)

type fakeCustomerRepository struct {
	customers []domain.Customer
}

func (r *fakeCustomerRepository) GetAll() []domain.Customer {
	return r.customers
}

func (r *fakeCustomerRepository) GetByID(id string) (*domain.Customer, error) {
	for _, customer := range r.customers {
		if customer.ID == id {
			return &customer, nil
		}
	}

	return nil, errors.New("customer not found")
}

func (r *fakeCustomerRepository) Create(customer domain.Customer) error {
	r.customers = append(r.customers, customer)
	return nil
}

func (r *fakeCustomerRepository) Update(id string, customer domain.Customer) error {
	return nil
}

func (r *fakeCustomerRepository) Delete(id string) error {
	return nil
}

type fakeCustomerProducer struct {
	publishedEvents []messaging.CustomerCreatedEvent
}

func (p *fakeCustomerProducer) PublishCustomerCreated(
	ctx context.Context,
	event messaging.CustomerCreatedEvent,
) error {
	p.publishedEvents = append(p.publishedEvents, event)
	return nil
}

func TestCustomerService_Create_ReturnsError_WhenFirstNameIsEmpty(t *testing.T) {
	repository := &fakeCustomerRepository{}
	producer := &fakeCustomerProducer{}
	service := NewCustomerService(repository, producer)

	_, err := service.Create(context.Background(), domain.Customer{
		LastName: "Burtin",
		Email:    "francisco@example.com",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "firstName is required" {
		t.Fatalf("expected firstName validation error, got %s", err.Error())
	}

	if len(producer.publishedEvents) != 0 {
		t.Fatalf("expected 0 published events, got %d", len(producer.publishedEvents))
	}
}

func TestCustomerService_Create_ReturnsError_WhenEmailIsInvalid(t *testing.T) {
	repository := &fakeCustomerRepository{}
	producer := &fakeCustomerProducer{}
	service := NewCustomerService(repository, producer)

	_, err := service.Create(context.Background(), domain.Customer{
		FirstName: "Francisco",
		LastName:  "Burtin",
		Email:     "invalid-email",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "email is invalid" {
		t.Fatalf("expected email validation error, got %s", err.Error())
	}

	if len(producer.publishedEvents) != 0 {
		t.Fatalf("expected 0 published events, got %d", len(producer.publishedEvents))
	}
}

func TestCustomerService_Create_CreatesCustomer_WhenValid(t *testing.T) {
	repository := &fakeCustomerRepository{}
	producer := &fakeCustomerProducer{}
	service := NewCustomerService(repository, producer)

	customer, err := service.Create(context.Background(), domain.Customer{
		FirstName: "Francisco",
		LastName:  "Burtin",
		Email:     "francisco@example.com",
	})

	if err != nil {
		t.Fatalf("expected nil error, got %s", err.Error())
	}

	if customer.ID == "" {
		t.Fatal("expected generated customer ID")
	}

	if customer.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}

	if len(repository.customers) != 1 {
		t.Fatalf("expected 1 customer in repository, got %d", len(repository.customers))
	}

	if len(producer.publishedEvents) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(producer.publishedEvents))
	}

	event := producer.publishedEvents[0]

	if event.EventType != "customer.created" {
		t.Fatalf("expected eventType customer.created, got %s", event.EventType)
	}

	if event.ID != customer.ID {
		t.Fatalf("expected event ID %s, got %s", customer.ID, event.ID)
	}

	if event.Email != customer.Email {
		t.Fatalf("expected event email %s, got %s", customer.Email, event.Email)
	}
}
