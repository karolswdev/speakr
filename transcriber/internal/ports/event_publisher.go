package ports

import (
	"context"
)

// Event represents a domain event
type Event struct {
	Subject string
	Data    interface{}
}

// EventPublisher defines the interface for publishing events
type EventPublisher interface {
	PublishEvent(ctx context.Context, event Event) error
}