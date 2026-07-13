package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/domain"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/messaging"
)

type fakeCustomerRepository struct {
	customers []domain.Customer

	createdCustomer domain.Customer
	createdOutbox   domain.OutboxEvent

	createWithOutboxCalled bool
	createWithOutboxError  error

	getByIDCustomer *domain.Customer
	getByIDError    error

	updateCalled bool
	updateID     string
	updated      domain.Customer
	updateError  error

	deleteCalled bool
	deleteID     string
	deleteError  error
}

func (r *fakeCustomerRepository) GetAll() []domain.Customer {
	return r.customers
}

func (r *fakeCustomerRepository) GetByID(
	id string,
) (*domain.Customer, error) {
	return r.getByIDCustomer, r.getByIDError
}

func (r *fakeCustomerRepository) CreateWithOutbox(
	ctx context.Context,
	customer domain.Customer,
	outboxEvent domain.OutboxEvent,
) error {
	r.createWithOutboxCalled = true
	r.createdCustomer = customer
	r.createdOutbox = outboxEvent

	return r.createWithOutboxError
}

func (r *fakeCustomerRepository) Update(
	id string,
	customer domain.Customer,
) error {
	r.updateCalled = true
	r.updateID = id
	r.updated = customer

	return r.updateError
}

func (r *fakeCustomerRepository) Delete(id string) error {
	r.deleteCalled = true
	r.deleteID = id

	return r.deleteError
}

func TestCustomerService_Create_SavesCustomerAndOutboxEvent(
	t *testing.T,
) {
	repository := &fakeCustomerRepository{}
	service := NewCustomerService(repository)

	input := domain.Customer{
		FirstName: "Francisco",
		LastName:  "Burtin",
		Email:     "francisco@example.com",
	}

	result, err := service.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !repository.createWithOutboxCalled {
		t.Fatal("expected CreateWithOutbox to be called")
	}

	if result.ID == "" {
		t.Fatal("expected customer ID to be generated")
	}

	if result.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be generated")
	}

	if repository.createdCustomer.ID != result.ID {
		t.Fatalf(
			"expected stored customer ID %q, got %q",
			result.ID,
			repository.createdCustomer.ID,
		)
	}

	if repository.createdCustomer.FirstName != input.FirstName {
		t.Fatalf(
			"expected first name %q, got %q",
			input.FirstName,
			repository.createdCustomer.FirstName,
		)
	}

	if repository.createdCustomer.LastName != input.LastName {
		t.Fatalf(
			"expected last name %q, got %q",
			input.LastName,
			repository.createdCustomer.LastName,
		)
	}

	if repository.createdCustomer.Email != input.Email {
		t.Fatalf(
			"expected email %q, got %q",
			input.Email,
			repository.createdCustomer.Email,
		)
	}

	outboxEvent := repository.createdOutbox

	if outboxEvent.ID == "" {
		t.Fatal("expected outbox event ID to be generated")
	}

	if outboxEvent.EventID == "" {
		t.Fatal("expected event ID to be generated")
	}

	if outboxEvent.AggregateID != result.ID {
		t.Fatalf(
			"expected aggregate ID %q, got %q",
			result.ID,
			outboxEvent.AggregateID,
		)
	}

	if outboxEvent.AggregateType != "customer" {
		t.Fatalf(
			"expected aggregate type customer, got %q",
			outboxEvent.AggregateType,
		)
	}

	if outboxEvent.EventType != "customer.created" {
		t.Fatalf(
			"expected event type customer.created, got %q",
			outboxEvent.EventType,
		)
	}

	if outboxEvent.Status != domain.OutboxPending {
		t.Fatalf(
			"expected status %q, got %q",
			domain.OutboxPending,
			outboxEvent.Status,
		)
	}

	if outboxEvent.Attempts != 0 {
		t.Fatalf(
			"expected zero attempts, got %d",
			outboxEvent.Attempts,
		)
	}

	var publishedEvent messaging.CustomerCreatedEvent

	if err := json.Unmarshal(
		outboxEvent.Payload,
		&publishedEvent,
	); err != nil {
		t.Fatalf("failed to decode outbox payload: %v", err)
	}

	if publishedEvent.EventID != outboxEvent.EventID {
		t.Fatalf(
			"expected payload event ID %q, got %q",
			outboxEvent.EventID,
			publishedEvent.EventID,
		)
	}

	if publishedEvent.EventType != "customer.created" {
		t.Fatalf(
			"expected payload event type customer.created, got %q",
			publishedEvent.EventType,
		)
	}

	if publishedEvent.ID != result.ID {
		t.Fatalf(
			"expected payload customer ID %q, got %q",
			result.ID,
			publishedEvent.ID,
		)
	}

	if publishedEvent.FirstName != input.FirstName {
		t.Fatalf(
			"expected payload first name %q, got %q",
			input.FirstName,
			publishedEvent.FirstName,
		)
	}

	if publishedEvent.LastName != input.LastName {
		t.Fatalf(
			"expected payload last name %q, got %q",
			input.LastName,
			publishedEvent.LastName,
		)
	}

	if publishedEvent.Email != input.Email {
		t.Fatalf(
			"expected payload email %q, got %q",
			input.Email,
			publishedEvent.Email,
		)
	}
}

func TestCustomerService_Create_ReturnsRepositoryError(
	t *testing.T,
) {
	expectedError := errors.New("transaction failed")

	repository := &fakeCustomerRepository{
		createWithOutboxError: expectedError,
	}

	service := NewCustomerService(repository)

	_, err := service.Create(
		context.Background(),
		domain.Customer{
			FirstName: "Francisco",
			LastName:  "Burtin",
			Email:     "francisco@example.com",
		},
	)

	if !errors.Is(err, expectedError) {
		t.Fatalf(
			"expected error %v, got %v",
			expectedError,
			err,
		)
	}

	if !repository.createWithOutboxCalled {
		t.Fatal("expected CreateWithOutbox to be called")
	}
}

func TestCustomerService_Create_ValidationErrors(
	t *testing.T,
) {
	tests := []struct {
		name          string
		customer      domain.Customer
		expectedError string
	}{
		{
			name: "first name is empty",
			customer: domain.Customer{
				FirstName: "",
				LastName:  "Burtin",
				Email:     "francisco@example.com",
			},
			expectedError: "firstName is required",
		},
		{
			name: "first name contains spaces only",
			customer: domain.Customer{
				FirstName: "   ",
				LastName:  "Burtin",
				Email:     "francisco@example.com",
			},
			expectedError: "firstName is required",
		},
		{
			name: "last name is empty",
			customer: domain.Customer{
				FirstName: "Francisco",
				LastName:  "",
				Email:     "francisco@example.com",
			},
			expectedError: "lastName is required",
		},
		{
			name: "email is empty",
			customer: domain.Customer{
				FirstName: "Francisco",
				LastName:  "Burtin",
				Email:     "",
			},
			expectedError: "email is required",
		},
		{
			name: "email is invalid",
			customer: domain.Customer{
				FirstName: "Francisco",
				LastName:  "Burtin",
				Email:     "invalid-email",
			},
			expectedError: "email is invalid",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeCustomerRepository{}
			service := NewCustomerService(repository)

			_, err := service.Create(
				context.Background(),
				test.customer,
			)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != test.expectedError {
				t.Fatalf(
					"expected error %q, got %q",
					test.expectedError,
					err.Error(),
				)
			}

			if repository.createWithOutboxCalled {
				t.Fatal(
					"expected repository not to be called",
				)
			}
		})
	}
}

func TestCustomerService_GetAll_ReturnsCustomers(
	t *testing.T,
) {
	expected := []domain.Customer{
		{
			ID:        "1",
			FirstName: "Francisco",
			LastName:  "Burtin",
			Email:     "francisco@example.com",
		},
	}

	repository := &fakeCustomerRepository{
		customers: expected,
	}

	service := NewCustomerService(repository)

	result := service.GetAll()

	if len(result) != 1 {
		t.Fatalf(
			"expected one customer, got %d",
			len(result),
		)
	}

	if result[0].ID != expected[0].ID {
		t.Fatalf(
			"expected customer ID %q, got %q",
			expected[0].ID,
			result[0].ID,
		)
	}
}

func TestCustomerService_GetByID_ReturnsCustomer(
	t *testing.T,
) {
	expected := &domain.Customer{
		ID:        "customer-1",
		FirstName: "Francisco",
		LastName:  "Burtin",
		Email:     "francisco@example.com",
	}

	repository := &fakeCustomerRepository{
		getByIDCustomer: expected,
	}

	service := NewCustomerService(repository)

	result, err := service.GetByID(expected.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.ID != expected.ID {
		t.Fatalf(
			"expected customer ID %q, got %q",
			expected.ID,
			result.ID,
		)
	}
}

func TestCustomerService_Update_ValidCustomer(
	t *testing.T,
) {
	repository := &fakeCustomerRepository{}
	service := NewCustomerService(repository)

	customer := domain.Customer{
		FirstName: "Francisco",
		LastName:  "Burtin",
		Email:     "francisco@example.com",
	}

	err := service.Update("customer-1", customer)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !repository.updateCalled {
		t.Fatal("expected Update to be called")
	}

	if repository.updateID != "customer-1" {
		t.Fatalf(
			"expected ID customer-1, got %q",
			repository.updateID,
		)
	}
}

func TestCustomerService_Update_InvalidCustomer(
	t *testing.T,
) {
	repository := &fakeCustomerRepository{}
	service := NewCustomerService(repository)

	err := service.Update(
		"customer-1",
		domain.Customer{
			FirstName: "",
			LastName:  "Burtin",
			Email:     "francisco@example.com",
		},
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if repository.updateCalled {
		t.Fatal("expected Update not to be called")
	}
}

func TestCustomerService_Delete_CallsRepository(
	t *testing.T,
) {
	repository := &fakeCustomerRepository{}
	service := NewCustomerService(repository)

	err := service.Delete("customer-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !repository.deleteCalled {
		t.Fatal("expected Delete to be called")
	}

	if repository.deleteID != "customer-1" {
		t.Fatalf(
			"expected customer-1, got %q",
			repository.deleteID,
		)
	}
}
