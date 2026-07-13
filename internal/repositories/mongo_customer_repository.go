package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MongoCustomerRepository struct {
	client             *mongo.Client
	customerCollection *mongo.Collection
	outboxCollection   *mongo.Collection
}

func NewMongoCustomerRepository(database *mongo.Database) *MongoCustomerRepository {
	return &MongoCustomerRepository{
		client:             database.Client(),
		customerCollection: database.Collection("customers"),
		outboxCollection:   database.Collection("outbox_events"),
	}
}

func (r *MongoCustomerRepository) GetAll() []domain.Customer {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.customerCollection.Find(ctx, bson.M{})
	if err != nil {
		return []domain.Customer{}
	}
	defer cursor.Close(ctx)

	var customers []domain.Customer
	if err := cursor.All(ctx, &customers); err != nil {
		return []domain.Customer{}
	}

	return customers
}

func (r *MongoCustomerRepository) GetByID(id string) (*domain.Customer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var customer domain.Customer

	err := r.customerCollection.FindOne(ctx, bson.M{"id": id}).Decode(&customer)
	if err != nil {
		return nil, errors.New("customer not found")
	}

	return &customer, nil
}

func (r *MongoCustomerRepository) CreateWithOutbox(
	ctx context.Context,
	customer domain.Customer,
	outboxEvent domain.OutboxEvent,
) error {
	session, err := r.client.StartSession()
	if err != nil {
		return fmt.Errorf("start MongoDB session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(
		ctx,
		func(sessionContext context.Context) (any, error) {
			if _, err := r.customerCollection.InsertOne(
				sessionContext,
				customer,
			); err != nil {
				return nil, fmt.Errorf("insert customer: %w", err)
			}

			if _, err := r.outboxCollection.InsertOne(
				sessionContext,
				outboxEvent,
			); err != nil {
				return nil, fmt.Errorf("insert outbox event: %w", err)
			}

			return nil, nil
		},
	)

	if err != nil {
		return fmt.Errorf("create customer with outbox: %w", err)
	}
	return nil
}

func (r *MongoCustomerRepository) Update(id string, customer domain.Customer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"firstName": customer.FirstName,
			"lastName":  customer.LastName,
			"email":     customer.Email,
		},
	}

	result, err := r.customerCollection.UpdateOne(ctx, bson.M{"id": id}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("customer not found")
	}

	return nil
}

func (r *MongoCustomerRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := r.customerCollection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("customer not found")
	}

	return nil
}
