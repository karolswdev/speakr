package ports

import (
	"context"
)

// EventHandler defines the function signature for handling events
type EventHandler func(ctx context.Context, subject string, data []byte) error

// EventSubscriber defines the interface for subscribing to events
type EventSubscriber interface {
	Subscribe(ctx context.Context, subject string, handler EventHandler) error
	Close() error
}