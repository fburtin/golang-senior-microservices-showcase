package services

import (
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

func (s *CustomerService) Create(customer domain.Customer) error {
	customer.ID = time.Now().Format("20060102150405")
	customer.CreatedAt = time.Now()
	return s.repository.Create(customer)
}
