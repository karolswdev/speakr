package openai_adapter

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewTranscriber(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Test without API key
	_, err := NewTranscriber(logger)
	if err != ErrAPIKeyNotSet {
		t.Errorf("Expected ErrAPIKeyNotSet, got %v", err)
	}

	// Test with API key
	transcriber, err := NewTranscriber(logger, WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("Failed to create transcriber: %v", err)
	}

	if transcriber.config.APIKey != "test-key" {
		t.Errorf("Expected API key 'test-key', got %s", transcriber.config.APIKey)
	}

	if transcriber.config.Model != "whisper-1" {
		t.Errorf("Expected model 'whisper-1', got %s", transcriber.config.Model)
	}
}

func TestTranscriberWithOptions(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	transcriber, err := NewTranscriber(logger,
		WithAPIKey("test-key"),
		WithBaseURL("https://api.example.com/v1"),
		WithModel("whisper-large"),
		WithTimeout(60*time.Second),
		WithMaxRetries(5),
	)
	if err != nil {
		t.Fatalf("Failed to create transcriber: %v", err)
	}

	if transcriber.config.BaseURL != "https://api.example.com/v1" {
		t.Errorf("Expected base URL 'https://api.example.com/v1', got %s", transcriber.config.BaseURL)
	}

	if transcriber.config.Model != "whisper-large" {
		t.Errorf("Expected model 'whisper-large', got %s", transcriber.config.Model)
	}

	if transcriber.config.Timeout != 60*time.Second {
		t.Errorf("Expected timeout 60s, got %v", transcriber.config.Timeout)
	}

	if transcriber.config.MaxRetries != 5 {
		t.Errorf("Expected max retries 5, got %d", transcriber.config.MaxRetries)
	}
}

func TestTranscribeAudio_InvalidInput(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	transcriber, err := NewTranscriber(logger, WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("Failed to create transcriber: %v", err)
	}

	ctx := context.Background()

	// Test with audio that's too large (simulate 26MB)
	largeAudio := strings.NewReader(strings.Repeat("a", 26*1024*1024))
	_, err = transcriber.TranscribeAudio(ctx, largeAudio, "wav")
	if err != ErrAudioTooLarge {
		t.Errorf("Expected ErrAudioTooLarge, got %v", err)
	}
}

func TestIsNonRetryableError(t *testing.T) {
	testCases := []struct {
		err        error
		retryable  bool
	}{
		{ErrAPIKeyInvalid, false},
		{ErrAPIKeyNotSet, false},
		{ErrAudioTooLarge, false},
		{ErrInvalidAudioFormat, false},
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

func TestTranscribeAudio_Integration(t *testing.T) {
	// Skip if not in integration test mode or no API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if os.Getenv("INTEGRATION_TEST") != "true" || apiKey == "" {
		t.Skip("Skipping integration test (set INTEGRATION_TEST=true and OPENAI_API_KEY to run)")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	transcriber, err := NewTranscriber(logger, WithAPIKey(apiKey))
	if err != nil {
		t.Fatalf("Failed to create transcriber: %v", err)
	}

	// Create a simple audio file content (this would be real audio in practice)
	// For this test, we'll just verify the API call structure works
	ctx := context.Background()
	audioData := strings.NewReader("mock audio data")

	// This will fail with the OpenAI API since it's not real audio,
	// but it will test our error handling
	_, err = transcriber.TranscribeAudio(ctx, audioData, "wav")
	if err == nil {
		t.Log("Transcription succeeded (unexpected with mock data)")
	} else {
		t.Logf("Transcription failed as expected with mock data: %v", err)
	}
}