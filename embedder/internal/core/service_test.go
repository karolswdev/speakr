package core

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"speakr/embedder/internal/ports"
)

// Mock implementations for testing
type mockEmbeddingGenerator struct {
	generateEmbeddingFunc func(ctx context.Context, text string) ([]float32, error)
}

func (m *mockEmbeddingGenerator) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if m.generateEmbeddingFunc != nil {
		return m.generateEmbeddingFunc(ctx, text)
	}
	// Return a mock embedding vector
	return []float32{0.1, 0.2, 0.3, 0.4, 0.5}, nil
}

type mockVectorStore struct {
	storeRecordFunc func(ctx context.Context, record ports.VectorRecord) error
	getRecordFunc   func(ctx context.Context, recordingID string) (*ports.VectorRecord, error)
}

func (m *mockVectorStore) StoreRecord(ctx context.Context, record ports.VectorRecord) error {
	if m.storeRecordFunc != nil {
		return m.storeRecordFunc(ctx, record)
	}
	return nil
}

func (m *mockVectorStore) GetRecord(ctx context.Context, recordingID string) (*ports.VectorRecord, error) {
	if m.getRecordFunc != nil {
		return m.getRecordFunc(ctx, recordingID)
	}
	return &ports.VectorRecord{
		RecordingID:     recordingID,
		TranscribedText: "mock transcribed text",
		Tags:            []string{"test"},
		Embedding:       []float32{0.1, 0.2, 0.3, 0.4, 0.5},
	}, nil
}

func createTestService() (*Service, *mockEmbeddingGenerator, *mockVectorStore) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	
	embeddingGenerator := &mockEmbeddingGenerator{}
	vectorStore := &mockVectorStore{}
	
	service := NewService(embeddingGenerator, vectorStore, logger)
	
	return service, embeddingGenerator, vectorStore
}

func TestService_ProcessTranscription(t *testing.T) {
	service, _, _ := createTestService()
	
	ctx := context.Background()
	recordingID := "test-recording-id"
	transcribedText := "This is a test transcription"
	tags := []string{"test", "embedding"}
	
	// Test successful processing
	err := service.ProcessTranscription(ctx, recordingID, transcribedText, tags)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestService_ProcessTranscription_EmptyText(t *testing.T) {
	service, _, _ := createTestService()
	
	ctx := context.Background()
	recordingID := "test-recording-id"
	transcribedText := ""
	tags := []string{"test"}
	
	// Test with empty text
	err := service.ProcessTranscription(ctx, recordingID, transcribedText, tags)
	if err != ErrEmptyText {
		t.Errorf("Expected ErrEmptyText, got %v", err)
	}
}

func TestService_HandleTranscriptionEvent(t *testing.T) {
	service, _, _ := createTestService()
	
	ctx := context.Background()
	subject := "speakr.event.transcription.succeeded"
	
	// Valid event data
	validEventData := `{
		"recording_id": "test-recording-id",
		"transcribed_text": "This is a test transcription",
		"tags": ["test", "embedding"],
		"metadata": {"source": "test"}
	}`
	
	err := service.HandleTranscriptionEvent(ctx, subject, []byte(validEventData))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestService_HandleTranscriptionEvent_InvalidJSON(t *testing.T) {
	service, _, _ := createTestService()
	
	ctx := context.Background()
	subject := "speakr.event.transcription.succeeded"
	
	// Invalid JSON
	invalidEventData := `{"invalid": json}`
	
	err := service.HandleTranscriptionEvent(ctx, subject, []byte(invalidEventData))
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestService_HandleTranscriptionEvent_MissingRecordingID(t *testing.T) {
	service, _, _ := createTestService()
	
	ctx := context.Background()
	subject := "speakr.event.transcription.succeeded"
	
	// Missing recording_id
	eventData := `{
		"transcribed_text": "This is a test transcription",
		"tags": ["test"]
	}`
	
	err := service.HandleTranscriptionEvent(ctx, subject, []byte(eventData))
	if err != ErrMissingRecordingID {
		t.Errorf("Expected ErrMissingRecordingID, got %v", err)
	}
}

func TestService_HandleTranscriptionEvent_EmptyText(t *testing.T) {
	service, _, _ := createTestService()
	
	ctx := context.Background()
	subject := "speakr.event.transcription.succeeded"
	
	// Empty transcribed_text
	eventData := `{
		"recording_id": "test-recording-id",
		"transcribed_text": "",
		"tags": ["test"]
	}`
	
	err := service.HandleTranscriptionEvent(ctx, subject, []byte(eventData))
	if err != ErrEmptyText {
		t.Errorf("Expected ErrEmptyText, got %v", err)
	}
}

func TestService_GetRecord(t *testing.T) {
	service, _, _ := createTestService()
	
	ctx := context.Background()
	recordingID := "test-recording-id"
	
	// Test successful retrieval
	record, err := service.GetRecord(ctx, recordingID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if record == nil {
		t.Fatal("Expected record, got nil")
	}
	
	if record.RecordingID != recordingID {
		t.Errorf("Expected recording ID %s, got %s", recordingID, record.RecordingID)
	}
}

func TestService_ErrorHandling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	
	// Test with failing embedding generator
	failingEmbedder := &mockEmbeddingGenerator{
		generateEmbeddingFunc: func(ctx context.Context, text string) ([]float32, error) {
			return nil, ErrEmptyText
		},
	}
	
	vectorStore := &mockVectorStore{}
	service := NewService(failingEmbedder, vectorStore, logger)
	
	ctx := context.Background()
	err := service.ProcessTranscription(ctx, "test-id", "test text", []string{"test"})
	if err == nil {
		t.Error("Expected error from failing embedder, got nil")
	}
}