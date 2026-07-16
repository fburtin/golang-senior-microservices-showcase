package domain

import "time"

type ProcessedEventStatus string

const (
	ProcessedEventProcessing ProcessedEventStatus = "PROCESSING"
	ProcessedEventCompleted  ProcessedEventStatus = "COMPLETED"
)

type ProcessedEvent struct {
	EventID     string               `bson:"eventId" json:"eventId"`
	EventType   string               `bson:"eventType" json:"eventType"`
	Status      ProcessedEventStatus `bson:"status" json:"status"`
	StartedAt   time.Time            `bson:"startedAt" json:"startedAt"`
	CompletedAt *time.Time           `bson:"completedAt,omitempty" json:"completedAt,omitempty"`
}
