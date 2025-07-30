package openai_adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// TranscriberConfig holds configuration for the OpenAI transcriber
type TranscriberConfig struct {
	APIKey     string
	BaseURL    string
	Model      string
	Timeout    time.Duration
	MaxRetries int
}

// TranscriberOption is a functional option for configuring the transcriber
type TranscriberOption func(*TranscriberConfig)

// WithAPIKey sets the OpenAI API key
func WithAPIKey(apiKey string) TranscriberOption {
	return func(c *TranscriberConfig) {
		c.APIKey = apiKey
	}
}

// WithBaseURL sets the OpenAI API base URL
func WithBaseURL(baseURL string) TranscriberOption {
	return func(c *TranscriberConfig) {
		c.BaseURL = baseURL
	}
}

// WithModel sets the transcription model
func WithModel(model string) TranscriberOption {
	return func(c *TranscriberConfig) {
		c.Model = model
	}
}

// WithTimeout sets the request timeout
func WithTimeout(timeout time.Duration) TranscriberOption {
	return func(c *TranscriberConfig) {
		c.Timeout = timeout
	}
}

// WithMaxRetries sets the maximum number of retries
func WithMaxRetries(maxRetries int) TranscriberOption {
	return func(c *TranscriberConfig) {
		c.MaxRetries = maxRetries
	}
}

// Transcriber implements the TranscriptionService port using OpenAI Whisper API
type Transcriber struct {
	config TranscriberConfig
	client *http.Client
	logger *slog.Logger
}

// NewTranscriber creates a new OpenAI transcriber
func NewTranscriber(logger *slog.Logger, opts ...TranscriberOption) (*Transcriber, error) {
	config := TranscriberConfig{
		BaseURL:    "https://api.openai.com/v1",
		Model:      "whisper-1",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	}

	for _, opt := range opts {
		opt(&config)
	}

	if config.APIKey == "" {
		return nil, ErrAPIKeyNotSet
	}

	// Validate base URL
	if err := validateBaseURL(config.BaseURL); err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &Transcriber{
		config: config,
		client: client,
		logger: logger,
	}, nil
}

// TranscribeAudio transcribes audio using OpenAI Whisper API
func (t *Transcriber) TranscribeAudio(ctx context.Context, audioData io.Reader, format string) (string, error) {
	logger := t.logger.With("model", t.config.Model, "format", format)

	logger.Info("Starting audio transcription")

	// Read audio data into memory
	audioBytes, err := io.ReadAll(audioData)
	if err != nil {
		logger.Error("Failed to read audio data", "error", err)
		return "", fmt.Errorf("failed to read audio data: %w", err)
	}

	// Check audio size (OpenAI has a 25MB limit)
	if len(audioBytes) > 25*1024*1024 {
		logger.Error("Audio file too large", "size", len(audioBytes))
		return "", ErrAudioTooLarge
	}

	// Attempt transcription with retries
	var lastErr error
	for attempt := 1; attempt <= t.config.MaxRetries; attempt++ {
		logger.Info("Transcription attempt", "attempt", attempt, "max_retries", t.config.MaxRetries)

		transcript, err := t.transcribeWithRetry(ctx, audioBytes, format)
		if err == nil {
			logger.Info("Transcription completed successfully", "text_length", len(transcript))
			return transcript, nil
		}

		lastErr = err
		logger.Warn("Transcription attempt failed", "attempt", attempt, "error", err)

		// Don't retry for certain error types
		if isNonRetryableError(err) {
			logger.Error("Non-retryable error encountered", "error", err)
			break
		}

		// Wait before retrying (exponential backoff)
		if attempt < t.config.MaxRetries {
			waitTime := time.Duration(attempt) * time.Second
			logger.Info("Waiting before retry", "wait_time", waitTime)
			time.Sleep(waitTime)
		}
	}

	logger.Error("All transcription attempts failed", "error", lastErr)
	return "", fmt.Errorf("transcription failed after %d attempts: %w", t.config.MaxRetries, lastErr)
}

// transcribeWithRetry performs a single transcription attempt
func (t *Transcriber) transcribeWithRetry(ctx context.Context, audioBytes []byte, format string) (string, error) {
	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add the audio file
	part, err := writer.CreateFormFile("file", fmt.Sprintf("audio.%s", format))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(audioBytes); err != nil {
		return "", fmt.Errorf("failed to write audio data: %w", err)
	}

	// Add model parameter
	if err := writer.WriteField("model", t.config.Model); err != nil {
		return "", fmt.Errorf("failed to write model field: %w", err)
	}

	// Add response format
	if err := writer.WriteField("response_format", "json"); err != nil {
		return "", fmt.Errorf("failed to write response format field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/audio/transcriptions", t.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+t.config.APIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Make the request
	resp, err := t.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return "", ErrRequestTimeout
		}
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Handle HTTP errors
	if resp.StatusCode != http.StatusOK {
		return "", t.handleHTTPError(resp.StatusCode, respBody)
	}

	// Parse successful response
	var transcriptionResp struct {
		Text string `json:"text"`
	}

	if err := json.Unmarshal(respBody, &transcriptionResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if transcriptionResp.Text == "" {
		return "", ErrEmptyTranscription
	}

	return strings.TrimSpace(transcriptionResp.Text), nil
}

// handleHTTPError converts HTTP errors to appropriate error types
func (t *Transcriber) handleHTTPError(statusCode int, body []byte) error {
	t.logger.Error("OpenAI API error", "status_code", statusCode, "response", string(body))

	switch statusCode {
	case http.StatusUnauthorized:
		return ErrAPIKeyInvalid
	case http.StatusTooManyRequests:
		return ErrQuotaExceeded
	case http.StatusBadRequest:
		// Try to parse error details
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
			} `json:"error"`
		}
		if json.Unmarshal(body, &errorResp) == nil {
			if strings.Contains(errorResp.Error.Message, "file size") {
				return ErrAudioTooLarge
			}
			if strings.Contains(errorResp.Error.Message, "format") {
				return ErrInvalidAudioFormat
			}
		}
		return fmt.Errorf("bad request: %s", string(body))
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return ErrServiceUnavailable
	default:
		return fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}
}

// isNonRetryableError determines if an error should not be retried
func isNonRetryableError(err error) bool {
	switch err {
	case ErrAPIKeyInvalid, ErrAPIKeyNotSet, ErrAudioTooLarge, ErrInvalidAudioFormat:
		return true
	default:
		return false
	}
}
// validateBaseURL validates the base URL format
func validateBaseURL(baseURL string) error {
	if baseURL == "" {
		return fmt.Errorf("base URL cannot be empty")
	}
	
	// Basic URL validation
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		return fmt.Errorf("base URL must start with http:// or https://")
	}
	
	return nil
}
