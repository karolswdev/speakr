package openai_adapter

import "errors"

// Custom error types for OpenAI-specific failures
var (
	ErrAPIKeyNotSet         = errors.New("OpenAI API key not set")
	ErrAPIKeyInvalid        = errors.New("OpenAI API key is invalid")
	ErrQuotaExceeded        = errors.New("OpenAI API quota exceeded")
	ErrAudioTooLarge        = errors.New("audio file exceeds OpenAI size limit (25MB)")
	ErrInvalidAudioFormat   = errors.New("invalid audio format for OpenAI API")
	ErrRequestTimeout       = errors.New("request to OpenAI API timed out")
	ErrServiceUnavailable   = errors.New("OpenAI API service unavailable")
	ErrEmptyTranscription   = errors.New("OpenAI API returned empty transcription")
	ErrNetworkError         = errors.New("network error communicating with OpenAI API")
)