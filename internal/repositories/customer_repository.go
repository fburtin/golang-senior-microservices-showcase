package repositories

import (
	"github.com/fburtin/golang-senior-microservices-showcase/internal/domain"
)

type CustomerRepository interface {
	GetAll() []domain.Customer
	GetByID(id string) (*domain.Customer, error)
	Create(customer domain.Customer) error
}
