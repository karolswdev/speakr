package nats_adapter

import (
	"context"
	"fmt"
	"log/slog"

	"speakr/embedder/internal/ports"

	"github.com/nats-io/nats.go"
)

// Subscriber implements the EventSubscriber port using NATS
type Subscriber struct {
	conn   *nats.Conn
	logger *slog.Logger
	subs   []*nats.Subscription
}

// NewSubscriber creates a new NATS subscriber
func NewSubscriber(conn *nats.Conn, logger *slog.Logger) *Subscriber {
	return &Subscriber{
		conn:   conn,
		logger: logger,
		subs:   make([]*nats.Subscription, 0),
	}
}

// Subscribe subscribes to a subject with the given handler
func (s *Subscriber) Subscribe(ctx context.Context, subject string, handler ports.EventHandler) error {
	logger := s.logger.With("subject", subject)

	logger.Info("Subscribing to NATS subject")

	// Create NATS message handler that wraps the port handler
	natsHandler := func(msg *nats.Msg) {
		// Create context for this message
		msgCtx := context.Background()
		
		// Add correlation ID if available from message headers
		if msg.Header != nil {
			if correlationID := msg.Header.Get("correlation_id"); correlationID != "" {
				msgCtx = context.WithValue(msgCtx, "correlation_id", correlationID)
			}
		}

		// Call the handler
		if err := handler(msgCtx, msg.Subject, msg.Data); err != nil {
			logger.Error("Handler failed to process message", 
				"error", err, 
				"subject", msg.Subject,
				"data_size", len(msg.Data))
		}
	}

	// Subscribe to the subject
	sub, err := s.conn.Subscribe(subject, natsHandler)
	if err != nil {
		logger.Error("Failed to subscribe to subject", "error", err)
		return fmt.Errorf("failed to subscribe to %s: %w", subject, err)
	}

	// Store subscription for cleanup
	s.subs = append(s.subs, sub)

	logger.Info("Successfully subscribed to subject")
	return nil
}

// Close unsubscribes from all subjects and cleans up
func (s *Subscriber) Close() error {
	s.logger.Info("Closing NATS subscriber", "subscriptions", len(s.subs))

	var lastErr error
	for _, sub := range s.subs {
		if err := sub.Unsubscribe(); err != nil {
			s.logger.Error("Failed to unsubscribe", "error", err)
			lastErr = err
		}
	}

	s.subs = nil
	return lastErr
}