package core

import "errors"

// Custom error types for predictable failures
var (
	ErrEmptyText          = errors.New("transcribed text cannot be empty")
	ErrMissingRecordingID = errors.New("recording ID is required")
	ErrRecordNotFound     = errors.New("vector record not found")
	ErrInvalidEvent       = errors.New("invalid event format")
)