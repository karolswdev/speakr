package openai_adapter

import "errors"

// Custom error types for OpenAI-specific failures
var (
	ErrAPIKeyNotSet       = errors.New("OpenAI API key not set")
	ErrAPIKeyInvalid      = errors.New("OpenAI API key is invalid")
	ErrQuotaExceeded      = errors.New("OpenAI API quota exceeded")
	ErrTextTooLong        = errors.New("text exceeds OpenAI token limit")
	ErrEmptyText          = errors.New("text cannot be empty")
	ErrRequestTimeout     = errors.New("request to OpenAI API timed out")
	ErrServiceUnavailable = errors.New("OpenAI API service unavailable")
	ErrEmptyEmbedding     = errors.New("OpenAI API returned empty embedding")
	ErrNetworkError       = errors.New("network error communicating with OpenAI API")
)