package customerdebt

import (
	"time"

	bcra "github.com/fburtin/golang-senior-microservices-showcase/internal/bcra"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Document struct {
	ID          bson.ObjectID   `bson:"_id,omitempty" json:"id"`
	CUIT        string          `bson:"cuit" json:"cuit"`
	BCRAStatus  int             `bson:"bcraStatus" json:"bcraStatus"`
	Result      bcra.DebtResult `bson:"result" json:"result"`
	RetrievedAt time.Time       `bson:"retrievedAt" json:"retrievedAt"`
	UpdatedAt   time.Time       `bson:"updatedAt" json:"updatedAt"`
}
