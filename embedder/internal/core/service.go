package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"speakr/embedder/internal/ports"

	"github.com/google/uuid"
)

// Service represents the core embedding service
type Service struct {
	embeddingGenerator ports.EmbeddingGenerator
	vectorStore        ports.VectorStore
	logger             *slog.Logger
}

// NewService creates a new embedding service
func NewService(
	embeddingGenerator ports.EmbeddingGenerator,
	vectorStore ports.VectorStore,
	logger *slog.Logger,
) *Service {
	return &Service{
		embeddingGenerator: embeddingGenerator,
		vectorStore:        vectorStore,
		logger:             logger,
	}
}

// TranscriptionSucceededEvent represents the transcription.succeeded event payload
type TranscriptionSucceededEvent struct {
	RecordingID     string                 `json:"recording_id"`
	TranscribedText string                 `json:"transcribed_text"`
	Tags            []string               `json:"tags"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ProcessTranscription handles a transcription.succeeded event
func (s *Service) ProcessTranscription(ctx context.Context, recordingID, transcribedText string, tags []string) error {
	correlationID := s.getCorrelationID(ctx)
	logger := s.logger.With(
		"correlation_id", correlationID,
		"recording_id", recordingID,
		"operation", "process_transcription",
	)

	logger.Info("Processing transcription for embedding", 
		"text_length", len(transcribedText), 
		"tags", tags)

	// Validate input
	if transcribedText == "" {
		logger.Error("Empty transcribed text provided")
		return ErrEmptyText
	}

	if len(transcribedText) > 8000 { // OpenAI embedding limit is ~8191 tokens
		logger.Warn("Transcribed text is very long, may exceed embedding limits", 
			"length", len(transcribedText))
	}

	// Generate embedding
	logger.Info("Generating embedding for transcribed text")
	embedding, err := s.embeddingGenerator.GenerateEmbedding(ctx, transcribedText)
	if err != nil {
		logger.Error("Failed to generate embedding", "error", err)
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	logger.Info("Embedding generated successfully", 
		"embedding_dimensions", len(embedding))

	// Create vector record
	record := ports.VectorRecord{
		RecordingID:     recordingID,
		TranscribedText: transcribedText,
		Tags:            tags,
		Embedding:       embedding,
	}

	// Store in vector database
	logger.Info("Storing vector record in database")
	if err := s.vectorStore.StoreRecord(ctx, record); err != nil {
		logger.Error("Failed to store vector record", "error", err)
		return fmt.Errorf("failed to store vector record: %w", err)
	}

	logger.Info("Transcription processed successfully", 
		"recording_id", recordingID,
		"embedding_dimensions", len(embedding),
		"tags_count", len(tags))

	return nil
}

// HandleTranscriptionEvent handles incoming transcription.succeeded events
func (s *Service) HandleTranscriptionEvent(ctx context.Context, subject string, data []byte) error {
	correlationID := s.getCorrelationID(ctx)
	logger := s.logger.With(
		"correlation_id", correlationID,
		"subject", subject,
		"operation", "handle_transcription_event",
	)

	logger.Info("Received transcription event", "data_size", len(data))

	// Parse the event
	var event TranscriptionSucceededEvent
	if err := json.Unmarshal(data, &event); err != nil {
		logger.Error("Failed to unmarshal transcription event", "error", err, "data", string(data))
		return fmt.Errorf("failed to unmarshal transcription event: %w", err)
	}

	// Validate event data
	if event.RecordingID == "" {
		logger.Error("Missing recording_id in transcription event")
		return ErrMissingRecordingID
	}

	if event.TranscribedText == "" {
		logger.Error("Missing transcribed_text in transcription event")
		return ErrEmptyText
	}

	// Add correlation ID to context for downstream processing
	ctx = context.WithValue(ctx, "correlation_id", correlationID)

	// Process the transcription
	return s.ProcessTranscription(ctx, event.RecordingID, event.TranscribedText, event.Tags)
}

// GetRecord retrieves a stored vector record
func (s *Service) GetRecord(ctx context.Context, recordingID string) (*ports.VectorRecord, error) {
	correlationID := s.getCorrelationID(ctx)
	logger := s.logger.With(
		"correlation_id", correlationID,
		"recording_id", recordingID,
		"operation", "get_record",
	)

	logger.Info("Retrieving vector record")

	record, err := s.vectorStore.GetRecord(ctx, recordingID)
	if err != nil {
		logger.Error("Failed to retrieve vector record", "error", err)
		return nil, fmt.Errorf("failed to retrieve vector record: %w", err)
	}

	if record == nil {
		logger.Warn("Vector record not found")
		return nil, ErrRecordNotFound
	}

	logger.Info("Vector record retrieved successfully", 
		"text_length", len(record.TranscribedText),
		"embedding_dimensions", len(record.Embedding),
		"tags_count", len(record.Tags))

	return record, nil
}

func (s *Service) getCorrelationID(ctx context.Context) string {
	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		if id, ok := correlationID.(string); ok {
			return id
		}
	}
	return uuid.New().String()
}