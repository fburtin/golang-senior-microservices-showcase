package messaging

import "time"

type CustomerCreatedEvent struct {
	EventType string    `json:"eventType"`
	ID        string    `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}
