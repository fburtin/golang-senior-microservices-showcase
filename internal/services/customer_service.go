package services

import (
	"errors"
	"strings"
	"time"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/domain"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/repositories"
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

func (s *CustomerService) Create(customer domain.Customer) (domain.Customer, error) {
	err := validateCustomer(customer)
	if err != nil {
		return domain.Customer{}, err
	}

	customer.ID = time.Now().Format("20060102150405")
	customer.CreatedAt = time.Now()

	err = s.repository.Create(customer)
	return customer, err
}

func (s *CustomerService) Delete(id string) error {
	return s.repository.Delete(id)
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
