package domain

import "time"

type OutboxStatus string

const (
	OutboxPending   OutboxStatus = "PENDING"
	OutboxPublished OutboxStatus = "PUBLISHED"
	OutboxFailed    OutboxStatus = "FAILED"
)

type OutboxEvent struct {
	ID            string       `bson:"id" json:"id"`
	EventID       string       `bson:"eventId" json:"eventId"`
	AggregateID   string       `bson:"aggregateId" json:"aggregateId"`
	AggregateType string       `bson:"aggregateType" json:"aggregateType"`
	EventType     string       `bson:"eventType" json:"eventType"`
	Payload       []byte       `bson:"payload" json:"payload"`
	Status        OutboxStatus `bson:"status" json:"status"`
	Attempts      int          `bson:"attempts" json:"attempts"`
	CreatedAt     time.Time    `bson:"createdAt" json:"createdAt"`
	PublishedAt   *time.Time   `bson:"publishedAt,omitempty" json:"publishedAt,omitempty"`
	LastError     string       `bson:"lastError,omitempty" json:"lastError,omitempty"`
}
