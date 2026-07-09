package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MongoCustomerRepository struct {
	collection *mongo.Collection
}

func NewMongoCustomerRepository(database *mongo.Database) *MongoCustomerRepository {
	return &MongoCustomerRepository{
		collection: database.Collection("customers"),
	}
}

func (r *MongoCustomerRepository) GetAll() []domain.Customer {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
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

	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&customer)
	if err != nil {
		return nil, errors.New("customer not found")
	}

	return &customer, nil
}

func (r *MongoCustomerRepository) Create(customer domain.Customer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, customer)
	return err
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

	result, err := r.collection.UpdateOne(ctx, bson.M{"id": id}, update)
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

	result, err := r.collection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("customer not found")
	}

	return nil
}
