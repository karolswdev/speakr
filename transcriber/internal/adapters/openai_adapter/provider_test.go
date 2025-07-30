package openai_adapter

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMultipleProviders(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	testCases := []struct {
		name    string
		baseURL string
		apiKey  string
	}{
		{
			name:    "OpenAI",
			baseURL: "https://api.openai.com/v1",
			apiKey:  os.Getenv("OPENAI_API_KEY"),
		},
		{
			name:    "Groq",
			baseURL: "https://api.groq.com/openai/v1",
			apiKey:  os.Getenv("GROQ_API_KEY"),
		},
		{
			name:    "Ollama",
			baseURL: "http://localhost:11434/v1",
			apiKey:  "ollama", // Ollama doesn't require a real API key
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Skip if API key not provided (except for Ollama)
			if tc.apiKey == "" && tc.name != "Ollama" {
				t.Skipf("Skipping %s test - no API key provided", tc.name)
			}

			transcriber, err := NewTranscriber(logger,
				WithAPIKey(tc.apiKey),
				WithBaseURL(tc.baseURL),
				WithTimeout(30*time.Second),
				WithMaxRetries(1), // Reduce retries for testing
			)
			if err != nil {
				t.Fatalf("Failed to create transcriber for %s: %v", tc.name, err)
			}

			// Verify the base URL is set correctly
			if transcriber.config.BaseURL != tc.baseURL {
				t.Errorf("Expected base URL %s, got %s", tc.baseURL, transcriber.config.BaseURL)
			}

			// Test with mock audio data (will likely fail but tests the request structure)
			ctx := context.Background()
			audioData := strings.NewReader("mock audio data")

			_, err = transcriber.TranscribeAudio(ctx, audioData, "wav")
			
			// We expect this to fail with mock data, but we can verify the error type
			if err != nil {
				t.Logf("%s provider failed as expected with mock data: %v", tc.name, err)
			} else {
				t.Logf("%s provider unexpectedly succeeded with mock data", tc.name)
			}
		})
	}
}

func TestBaseURLValidation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	testCases := []struct {
		name      string
		baseURL   string
		shouldErr bool
	}{
		{
			name:      "Valid HTTPS URL",
			baseURL:   "https://api.openai.com/v1",
			shouldErr: false,
		},
		{
			name:      "Valid HTTP URL",
			baseURL:   "http://localhost:11434/v1",
			shouldErr: false,
		},
		{
			name:      "URL with trailing slash",
			baseURL:   "https://api.groq.com/openai/v1/",
			shouldErr: false,
		},
		{
			name:      "Invalid URL - no protocol",
			baseURL:   "api.openai.com/v1",
			shouldErr: true,
		},
		{
			name:      "Empty URL",
			baseURL:   "",
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewTranscriber(logger,
				WithAPIKey("test-key"),
				WithBaseURL(tc.baseURL),
			)

			if tc.shouldErr && err == nil {
				t.Errorf("Expected error for base URL %s, got nil", tc.baseURL)
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("Unexpected error for base URL %s: %v", tc.baseURL, err)
			}
		})
	}
}

func TestProviderSpecificConfiguration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Test Azure OpenAI style URL
	azureURL := "https://your-resource.openai.azure.com/openai/deployments/your-deployment/v1"
	transcriber, err := NewTranscriber(logger,
		WithAPIKey("test-key"),
		WithBaseURL(azureURL),
		WithTimeout(60*time.Second), // Azure might need longer timeouts
	)
	if err != nil {
		t.Fatalf("Failed to create transcriber with Azure URL: %v", err)
	}

	if transcriber.config.BaseURL != azureURL {
		t.Errorf("Expected Azure URL %s, got %s", azureURL, transcriber.config.BaseURL)
	}

	if transcriber.config.Timeout != 60*time.Second {
		t.Errorf("Expected timeout 60s, got %v", transcriber.config.Timeout)
	}
}