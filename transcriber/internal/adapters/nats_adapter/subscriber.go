package nats_adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"speakr/transcriber/internal/core"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

// Subscriber handles NATS message subscriptions
type Subscriber struct {
	conn    *nats.Conn
	service *core.Service
	logger  *slog.Logger
}

// NewSubscriber creates a new NATS subscriber
func NewSubscriber(conn *nats.Conn, service *core.Service, logger *slog.Logger) *Subscriber {
	return &Subscriber{
		conn:    conn,
		service: service,
		logger:  logger,
	}
}

// Subscribe sets up subscriptions for all command subjects
func (s *Subscriber) Subscribe(ctx context.Context) error {
	subjects := []string{
		"speakr.command.recording.start",
		"speakr.command.recording.stop",
		"speakr.command.recording.cancel",
		"speakr.command.transcription.run",
	}

	for _, subject := range subjects {
		if _, err := s.conn.Subscribe(subject, s.handleMessage); err != nil {
			return fmt.Errorf("failed to subscribe to %s: %w", subject, err)
		}
		s.logger.Info("Subscribed to subject", "subject", subject)
	}

	return nil
}

// handleMessage processes incoming NATS messages
func (s *Subscriber) handleMessage(msg *nats.Msg) {
	correlationID := uuid.New().String()
	ctx := context.WithValue(context.Background(), "correlation_id", correlationID)
	
	logger := s.logger.With(
		"correlation_id", correlationID,
		"subject", msg.Subject,
	)

	logger.Info("Received message", "data", string(msg.Data))

	var err error
	switch msg.Subject {
	case "speakr.command.recording.start":
		err = s.handleStartRecording(ctx, msg.Data)
	case "speakr.command.recording.stop":
		err = s.handleStopRecording(ctx, msg.Data)
	case "speakr.command.recording.cancel":
		err = s.handleCancelRecording(ctx, msg.Data)
	case "speakr.command.transcription.run":
		err = s.handleTranscription(ctx, msg.Data)
	default:
		logger.Error("Unknown subject", "subject", msg.Subject)
		return
	}

	if err != nil {
		logger.Error("Failed to handle message", "error", err)
	} else {
		logger.Info("Message handled successfully")
	}
}

func (s *Subscriber) handleStartRecording(ctx context.Context, data []byte) error {
	var cmd core.StartRecordingCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return fmt.Errorf("failed to unmarshal start recording command: %w", err)
	}

	return s.service.StartRecording(ctx, cmd)
}

func (s *Subscriber) handleStopRecording(ctx context.Context, data []byte) error {
	var cmd core.StopRecordingCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return fmt.Errorf("failed to unmarshal stop recording command: %w", err)
	}

	return s.service.StopRecording(ctx, cmd)
}

func (s *Subscriber) handleCancelRecording(ctx context.Context, data []byte) error {
	var cmd core.CancelRecordingCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return fmt.Errorf("failed to unmarshal cancel recording command: %w", err)
	}

	return s.service.CancelRecording(ctx, cmd)
}

func (s *Subscriber) handleTranscription(ctx context.Context, data []byte) error {
	var cmd core.TranscriptionCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return fmt.Errorf("failed to unmarshal transcription command: %w", err)
	}

	return s.service.TranscribeAudio(ctx, cmd)
}