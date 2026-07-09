package services

import (
	"errors"
	"testing"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/domain"
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

func TestCustomerService_Create_ReturnsError_WhenFirstNameIsEmpty(t *testing.T) {
	repository := &fakeCustomerRepository{}
	service := NewCustomerService(repository)

	_, err := service.Create(domain.Customer{
		LastName: "Burtin",
		Email:    "francisco@example.com",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "firstName is required" {
		t.Fatalf("expected firstName validation error, got %s", err.Error())
	}
}

func TestCustomerService_Create_ReturnsError_WhenEmailIsInvalid(t *testing.T) {
	repository := &fakeCustomerRepository{}
	service := NewCustomerService(repository)

	_, err := service.Create(domain.Customer{
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
}

func TestCustomerService_Create_CreatesCustomer_WhenValid(t *testing.T) {
	repository := &fakeCustomerRepository{}
	service := NewCustomerService(repository)

	customer, err := service.Create(domain.Customer{
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
}
