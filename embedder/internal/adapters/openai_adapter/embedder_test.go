package openai_adapter

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestNewEmbedder(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Test without API key
	_, err := NewEmbedder(logger)
	if err != ErrAPIKeyNotSet {
		t.Errorf("Expected ErrAPIKeyNotSet, got %v", err)
	}

	// Test with API key
	embedder, err := NewEmbedder(logger, WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("Failed to create embedder: %v", err)
	}

	if embedder.config.APIKey != "test-key" {
		t.Errorf("Expected API key 'test-key', got %s", embedder.config.APIKey)
	}

	if embedder.config.Model != "text-embedding-ada-002" {
		t.Errorf("Expected model 'text-embedding-ada-002', got %s", embedder.config.Model)
	}
}

func TestEmbedderWithOptions(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	embedder, err := NewEmbedder(logger,
		WithAPIKey("test-key"),
		WithBaseURL("https://api.example.com/v1"),
		WithModel("text-embedding-3-small"),
		WithTimeout(60*time.Second),
		WithMaxRetries(5),
	)
	if err != nil {
		t.Fatalf("Failed to create embedder: %v", err)
	}

	if embedder.config.BaseURL != "https://api.example.com/v1" {
		t.Errorf("Expected base URL 'https://api.example.com/v1', got %s", embedder.config.BaseURL)
	}

	if embedder.config.Model != "text-embedding-3-small" {
		t.Errorf("Expected model 'text-embedding-3-small', got %s", embedder.config.Model)
	}

	if embedder.config.Timeout != 60*time.Second {
		t.Errorf("Expected timeout 60s, got %v", embedder.config.Timeout)
	}

	if embedder.config.MaxRetries != 5 {
		t.Errorf("Expected max retries 5, got %d", embedder.config.MaxRetries)
	}
}

func TestGenerateEmbedding_InvalidInput(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	embedder, err := NewEmbedder(logger, WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("Failed to create embedder: %v", err)
	}

	ctx := context.Background()

	// Test with empty text
	_, err = embedder.GenerateEmbedding(ctx, "")
	if err != ErrEmptyText {
		t.Errorf("Expected ErrEmptyText, got %v", err)
	}
}

func TestIsNonRetryableError(t *testing.T) {
	testCases := []struct {
		err        error
		retryable  bool
	}{
		{ErrAPIKeyInvalid, false},
		{ErrAPIKeyNotSet, false},
		{ErrTextTooLong, false},
		{ErrEmptyText, false},
		{ErrQuotaExceeded, true},
		{ErrServiceUnavailable, true},
		{ErrRequestTimeout, true},
	}

	for _, tc := range testCases {
		result := isNonRetryableError(tc.err)
		expected := !tc.retryable
		if result != expected {
			t.Errorf("For error %v, expected non-retryable=%v, got %v", tc.err, expected, result)
		}
	}
}

func TestGenerateEmbedding_Integration(t *testing.T) {
	// Skip if not in integration test mode or no API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if os.Getenv("INTEGRATION_TEST") != "true" || apiKey == "" {
		t.Skip("Skipping integration test (set INTEGRATION_TEST=true and OPENAI_API_KEY to run)")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	embedder, err := NewEmbedder(logger, WithAPIKey(apiKey))
	if err != nil {
		t.Fatalf("Failed to create embedder: %v", err)
	}

	ctx := context.Background()
	text := "This is a test sentence for generating embeddings."

	embedding, err := embedder.GenerateEmbedding(ctx, text)
	if err != nil {
		t.Fatalf("Failed to generate embedding: %v", err)
	}

	if len(embedding) == 0 {
		t.Error("Expected non-empty embedding")
	}

	// OpenAI text-embedding-ada-002 returns 1536 dimensions
	if len(embedding) != 1536 {
		t.Errorf("Expected 1536 dimensions, got %d", len(embedding))
	}

	t.Logf("Generated embedding with %d dimensions", len(embedding))
}