package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/domain"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/messaging"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/repositories"
)

type CustomerEventProducer interface {
	PublishCustomerCreated(
		ctx context.Context,
		event messaging.CustomerCreatedEvent,
	) error
}

type CustomerService struct {
	repository repositories.CustomerRepository
	producer   CustomerEventProducer
}

func NewCustomerService(repository repositories.CustomerRepository, producer CustomerEventProducer) *CustomerService {
	return &CustomerService{
		repository: repository,
		producer:   producer,
	}
}

func (s *CustomerService) GetByID(id string) (*domain.Customer, error) {
	return s.repository.GetByID(id)
}

func (s *CustomerService) GetAll() []domain.Customer {
	return s.repository.GetAll()
}

func (s *CustomerService) Create(ctx context.Context, customer domain.Customer) (domain.Customer, error) {
	err := validateCustomer(customer)
	if err != nil {
		return domain.Customer{}, err
	}

	customer.ID = time.Now().Format("20060102150405")
	customer.CreatedAt = time.Now()

	err = s.repository.Create(customer)

	if err != nil {
		return domain.Customer{}, err
	}

	// publish Kafka event here
	event := messaging.CustomerCreatedEvent{
		EventType: "customer.created",
		ID:        customer.ID,
		FirstName: customer.FirstName,
		LastName:  customer.LastName,
		Email:     customer.Email,
		CreatedAt: customer.CreatedAt,
	}

	err = s.producer.PublishCustomerCreated(ctx, event)
	if err != nil {
		return domain.Customer{}, err
	}

	return customer, err
}

func (s *CustomerService) Delete(id string) error {
	return s.repository.Delete(id)
}

func (s *CustomerService) Update(id string, customer domain.Customer) error {
	err := validateCustomer(customer)
	if err != nil {
		return err
	}

	return s.repository.Update(id, customer)
}

func validateCustomer(customer domain.Customer) error {
	if strings.TrimSpace(customer.FirstName) == "" {
		return errors.New("firstName is required")
	}

	if strings.TrimSpace(customer.LastName) == "" {
		return errors.New("lastName is required")
	}

	if strings.TrimSpace(customer.Email) == "" {
		return errors.New("email is required")
	}

	if !strings.Contains(customer.Email, "@") {
		return errors.New("email is invalid")
	}

	return nil
}
