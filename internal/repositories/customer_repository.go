package repositories

import (
	"context"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/domain"
)

type CustomerRepository interface {
	GetAll() []domain.Customer
	GetByID(id string) (*domain.Customer, error)
	CreateWithOutbox(
		ctx context.Context,
		customer domain.Customer,
		outboxEvent domain.OutboxEvent,
	) error
	Update(id string, customer domain.Customer) error
	Delete(id string) error
}
