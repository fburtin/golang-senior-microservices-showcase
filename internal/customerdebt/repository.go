package customerdebt

import (
	"context"
	"fmt"
	"time"

	bcra "github.com/fburtin/golang-senior-microservices-showcase/internal/bcra"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository interface {
	Upsert(
		ctx context.Context,
		cuit string,
		result bcra.DebtResult,
	) error

	FindByCUIT(
		ctx context.Context,
		cuit string,
	) (*Document, error)
}

type MongoRepository struct {
	collection *mongo.Collection
}

func NewMongoRepository(
	collection *mongo.Collection,
) *MongoRepository {
	return &MongoRepository{
		collection: collection,
	}
}

func (r *MongoRepository) Upsert(
	ctx context.Context,
	cuit string,
	result bcra.DebtResult,
) error {
	now := time.Now().UTC()

	filter := bson.M{
		"cuit": cuit,
	}

	update := bson.M{
		"$set": bson.M{
			"bcraStatus": 200,
			"result":     result,
			"updatedAt":  now,
		},
		"$setOnInsert": bson.M{
			"cuit":        cuit,
			"retrievedAt": now,
		},
	}

	_, err := r.collection.UpdateOne(
		ctx,
		filter,
		update,
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("upsert customer debt: %w", err)
	}

	return nil
}

func (r *MongoRepository) FindByCUIT(
	ctx context.Context,
	cuit string,
) (*Document, error) {
	var document Document

	err := r.collection.FindOne(
		ctx,
		bson.M{"cuit": cuit},
	).Decode(&document)

	if err != nil {
		return nil, fmt.Errorf("find customer debt by CUIT: %w", err)
	}

	return &document, nil
}

func (r *MongoRepository) CreateIndexes(
	ctx context.Context,
) error {
	index := mongo.IndexModel{
		Keys: bson.D{
			{Key: "cuit", Value: 1},
		},
		Options: options.Index().
			SetUnique(true),
	}

	_, err := r.collection.Indexes().CreateOne(
		ctx,
		index,
	)
	if err != nil {
		return fmt.Errorf(
			"create customer debt indexes: %w",
			err,
		)
	}

	return nil
}
