package openai_adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// EmbedderConfig holds configuration for the OpenAI embedder
type EmbedderConfig struct {
	APIKey     string
	BaseURL    string
	Model      string
	Timeout    time.Duration
	MaxRetries int
}

// EmbedderOption is a functional option for configuring the embedder
type EmbedderOption func(*EmbedderConfig)

// WithAPIKey sets the OpenAI API key
func WithAPIKey(apiKey string) EmbedderOption {
	return func(c *EmbedderConfig) {
		c.APIKey = apiKey
	}
}

// WithBaseURL sets the OpenAI API base URL
func WithBaseURL(baseURL string) EmbedderOption {
	return func(c *EmbedderConfig) {
		c.BaseURL = baseURL
	}
}

// WithModel sets the embedding model
func WithModel(model string) EmbedderOption {
	return func(c *EmbedderConfig) {
		c.Model = model
	}
}

// WithTimeout sets the request timeout
func WithTimeout(timeout time.Duration) EmbedderOption {
	return func(c *EmbedderConfig) {
		c.Timeout = timeout
	}
}

// WithMaxRetries sets the maximum number of retries
func WithMaxRetries(maxRetries int) EmbedderOption {
	return func(c *EmbedderConfig) {
		c.MaxRetries = maxRetries
	}
}

// Embedder implements the EmbeddingGenerator port using OpenAI Embeddings API
type Embedder struct {
	config EmbedderConfig
	client *http.Client
	logger *slog.Logger
}

// NewEmbedder creates a new OpenAI embedder
func NewEmbedder(logger *slog.Logger, opts ...EmbedderOption) (*Embedder, error) {
	config := EmbedderConfig{
		BaseURL:    "https://api.openai.com/v1",
		Model:      "text-embedding-ada-002",
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

	return &Embedder{
		config: config,
		client: client,
		logger: logger,
	}, nil
}

// GenerateEmbedding generates a vector embedding for the given text
func (e *Embedder) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	logger := e.logger.With("model", e.config.Model, "text_length", len(text))

	logger.Info("Generating embedding")

	// Validate input
	if text == "" {
		logger.Error("Empty text provided")
		return nil, ErrEmptyText
	}

	// Check text length (OpenAI has token limits)
	if len(text) > 8000 {
		logger.Warn("Text is very long, may exceed token limits", "length", len(text))
	}

	// Attempt embedding generation with retries
	var lastErr error
	for attempt := 1; attempt <= e.config.MaxRetries; attempt++ {
		logger.Info("Embedding attempt", "attempt", attempt, "max_retries", e.config.MaxRetries)

		embedding, err := e.generateWithRetry(ctx, text)
		if err == nil {
			logger.Info("Embedding generated successfully", "dimensions", len(embedding))
			return embedding, nil
		}

		lastErr = err
		logger.Warn("Embedding attempt failed", "attempt", attempt, "error", err)

		// Don't retry for certain error types
		if isNonRetryableError(err) {
			logger.Error("Non-retryable error encountered", "error", err)
			break
		}

		// Wait before retrying (exponential backoff)
		if attempt < e.config.MaxRetries {
			waitTime := time.Duration(attempt) * time.Second
			logger.Info("Waiting before retry", "wait_time", waitTime)
			time.Sleep(waitTime)
		}
	}

	logger.Error("All embedding attempts failed", "error", lastErr)
	return nil, fmt.Errorf("embedding generation failed after %d attempts: %w", e.config.MaxRetries, lastErr)
}

// generateWithRetry performs a single embedding generation attempt
func (e *Embedder) generateWithRetry(ctx context.Context, text string) ([]float32, error) {
	// Create request payload
	requestBody := map[string]interface{}{
		"input": text,
		"model": e.config.Model,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/embeddings", e.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+e.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := e.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ErrRequestTimeout
		}
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, e.handleHTTPError(resp.StatusCode, respBody)
	}

	// Parse successful response
	var embeddingResp struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(embeddingResp.Data) == 0 || len(embeddingResp.Data[0].Embedding) == 0 {
		return nil, ErrEmptyEmbedding
	}

	return embeddingResp.Data[0].Embedding, nil
}

// handleHTTPError converts HTTP errors to appropriate error types
func (e *Embedder) handleHTTPError(statusCode int, body []byte) error {
	e.logger.Error("OpenAI API error", "status_code", statusCode, "response", string(body))

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
			if strings.Contains(errorResp.Error.Message, "token") {
				return ErrTextTooLong
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
	case ErrAPIKeyInvalid, ErrAPIKeyNotSet, ErrTextTooLong, ErrEmptyText:
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