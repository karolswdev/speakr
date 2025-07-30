package nats_adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"speakr/transcriber/internal/ports"

	"github.com/nats-io/nats.go"
)

// Publisher handles NATS event publishing
type Publisher struct {
	conn   *nats.Conn
	logger *slog.Logger
}

// NewPublisher creates a new NATS publisher
func NewPublisher(conn *nats.Conn, logger *slog.Logger) *Publisher {
	return &Publisher{
		conn:   conn,
		logger: logger,
	}
}

// PublishEvent publishes an event to NATS
func (p *Publisher) PublishEvent(ctx context.Context, event ports.Event) error {
	correlationID := ctx.Value("correlation_id")
	logger := p.logger.With(
		"correlation_id", correlationID,
		"subject", event.Subject,
	)

	data, err := json.Marshal(event.Data)
	if err != nil {
		logger.Error("Failed to marshal event data", "error", err)
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	if err := p.conn.Publish(event.Subject, data); err != nil {
		logger.Error("Failed to publish event", "error", err)
		return fmt.Errorf("failed to publish event: %w", err)
	}

	logger.Info("Event published successfully", "data", string(data))
	return nil
}