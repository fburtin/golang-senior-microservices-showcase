package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/domain"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/messaging"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/repositories"
	"github.com/google/uuid"
)

type CustomerService struct {
	repository repositories.CustomerRepository
}

func NewCustomerService(repository repositories.CustomerRepository) *CustomerService {
	return &CustomerService{
		repository: repository,
	}
}

func (s *CustomerService) GetByID(id string) (*domain.Customer, error) {
	return s.repository.GetByID(id)
}

func (s *CustomerService) GetAll() []domain.Customer {
	return s.repository.GetAll()
}

func (s *CustomerService) Create(
	ctx context.Context,
	customer domain.Customer,
) (domain.Customer, error) {
	if err := validateCustomer(customer); err != nil {
		return domain.Customer{}, err
	}

	now := time.Now().UTC()

	customer.ID = now.Format("20060102150405")
	customer.CreatedAt = now

	event := messaging.CustomerCreatedEvent{
		EventID:   generateEventID(),
		EventType: "customer.created",
		ID:        customer.ID,
		FirstName: customer.FirstName,
		LastName:  customer.LastName,
		Email:     customer.Email,
		CreatedAt: customer.CreatedAt,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return domain.Customer{}, fmt.Errorf(
			"marshal customer-created event: %w",
			err,
		)
	}

	outboxEvent := domain.OutboxEvent{
		ID:            generateEventID(),
		EventID:       event.EventID,
		AggregateID:   customer.ID,
		AggregateType: "customer",
		EventType:     event.EventType,
		Payload:       payload,
		Status:        domain.OutboxPending,
		Attempts:      0,
		CreatedAt:     now,
	}

	if err := s.repository.CreateWithOutbox(
		ctx,
		customer,
		outboxEvent,
	); err != nil {
		return domain.Customer{}, err
	}

	return customer, nil
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

func generateEventID() string {
	return uuid.NewString()
}
