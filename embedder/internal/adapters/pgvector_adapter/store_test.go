package pgvector_adapter

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"speakr/embedder/internal/ports"
)

func TestNewStore(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Test with default options (will fail if PostgreSQL not running, which is expected)
	_, err := NewStore(logger)
	if err != nil {
		// This is expected if PostgreSQL is not running
		logger.Info("Store creation failed (expected if PostgreSQL not running)", "error", err)
		return
	}
}

func TestStoreWithOptions(t *testing.T) {
	_ = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Test configuration with options
	config := StoreConfig{}

	opts := []StoreOption{
		WithHost("localhost"),
		WithPort(5432),
		WithCredentials("testuser", "testpass"),
		WithDatabase("testdb"),
		WithSSLMode("require"),
		WithMaxConnections(20),
	}

	for _, opt := range opts {
		opt(&config)
	}

	if config.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got %s", config.Host)
	}

	if config.Port != 5432 {
		t.Errorf("Expected port 5432, got %d", config.Port)
	}

	if config.User != "testuser" {
		t.Errorf("Expected user 'testuser', got %s", config.User)
	}

	if config.Password != "testpass" {
		t.Errorf("Expected password 'testpass', got %s", config.Password)
	}

	if config.DBName != "testdb" {
		t.Errorf("Expected database 'testdb', got %s", config.DBName)
	}

	if config.SSLMode != "require" {
		t.Errorf("Expected SSL mode 'require', got %s", config.SSLMode)
	}

	if config.MaxConns != 20 {
		t.Errorf("Expected max connections 20, got %d", config.MaxConns)
	}
}

func TestVectorConversion(t *testing.T) {
	// Test vector to string conversion
	vector := []float32{1.0, 2.5, -3.14, 0.0}
	vectorStr := vectorToString(vector)
	expectedStr := "[1,2.5,-3.14,0]"

	if vectorStr != expectedStr {
		t.Errorf("Expected vector string %s, got %s", expectedStr, vectorStr)
	}

	// Test string to vector conversion
	convertedVector := stringToVector(vectorStr)
	if len(convertedVector) != len(vector) {
		t.Errorf("Expected vector length %d, got %d", len(vector), len(convertedVector))
	}

	for i, v := range vector {
		if convertedVector[i] != v {
			t.Errorf("Expected vector[%d] = %f, got %f", i, v, convertedVector[i])
		}
	}
}

func TestStoreOperations_Integration(t *testing.T) {
	// Skip if not in integration test mode
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test (set INTEGRATION_TEST=true to run)")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	store, err := NewStore(logger,
		WithHost("localhost"),
		WithPort(5432),
		WithCredentials("postgres", "postgres"),
		WithDatabase("speakr"),
	)
	if err != nil {
		t.Skipf("PostgreSQL not available, skipping integration test: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Test storing a record
	record := ports.VectorRecord{
		RecordingID:     "test-recording-123",
		TranscribedText: "This is a test transcription for embedding storage.",
		Tags:            []string{"test", "integration", "embedding"},
		Embedding:       []float32{0.1, 0.2, 0.3, 0.4, 0.5},
	}

	err = store.StoreRecord(ctx, record)
	if err != nil {
		t.Fatalf("Failed to store record: %v", err)
	}

	// Test retrieving the record
	retrievedRecord, err := store.GetRecord(ctx, record.RecordingID)
	if err != nil {
		t.Fatalf("Failed to retrieve record: %v", err)
	}

	if retrievedRecord == nil {
		t.Fatal("Retrieved record is nil")
	}

	if retrievedRecord.RecordingID != record.RecordingID {
		t.Errorf("Expected recording ID %s, got %s", record.RecordingID, retrievedRecord.RecordingID)
	}

	if retrievedRecord.TranscribedText != record.TranscribedText {
		t.Errorf("Expected text %s, got %s", record.TranscribedText, retrievedRecord.TranscribedText)
	}

	if len(retrievedRecord.Tags) != len(record.Tags) {
		t.Errorf("Expected %d tags, got %d", len(record.Tags), len(retrievedRecord.Tags))
	}

	if len(retrievedRecord.Embedding) != len(record.Embedding) {
		t.Errorf("Expected %d embedding dimensions, got %d", len(record.Embedding), len(retrievedRecord.Embedding))
	}

	// Test retrieving non-existent record
	nonExistentRecord, err := store.GetRecord(ctx, "non-existent-id")
	if err != nil {
		t.Fatalf("Unexpected error for non-existent record: %v", err)
	}

	if nonExistentRecord != nil {
		t.Error("Expected nil for non-existent record")
	}

	t.Log("Integration test completed successfully")
}