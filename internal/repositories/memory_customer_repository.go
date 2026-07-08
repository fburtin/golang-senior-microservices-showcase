package repositories

import (
	"errors"
	"time"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/domain"
)

type MemoryCustomerRepository struct {
	customers []domain.Customer
}

func NewMemoryCustomerRepository() *MemoryCustomerRepository {
	return &MemoryCustomerRepository{
		customers: []domain.Customer{
			{
				ID:        "1",
				FirstName: "Francisco",
				LastName:  "Burtin",
				Email:     "francisco@example.com",
				CreatedAt: time.Now(),
			},
		},
	}
}

func (r *MemoryCustomerRepository) GetAll() []domain.Customer {
	return r.customers
}

func (r *MemoryCustomerRepository) GetByID(id string) (*domain.Customer, error) {
	for _, customer := range r.customers {
		if customer.ID == id {
			return &customer, nil
		}
	}
	return nil, errors.New("customer not found")
}

func (r *MemoryCustomerRepository) Create(customer domain.Customer) error {
	customer.CreatedAt = time.Now()
	r.customers = append(r.customers, customer)
	return nil
}
