package pgvector_adapter

import "errors"

// Custom error types for database-specific failures
var (
	ErrDatabaseUnavailable = errors.New("database is unavailable")
	ErrMissingRecordingID  = errors.New("recording ID is required")
	ErrEmptyText           = errors.New("transcribed text cannot be empty")
	ErrEmptyEmbedding      = errors.New("embedding cannot be empty")
	ErrInvalidData         = errors.New("invalid data provided")
	ErrConnectionFailed    = errors.New("failed to connect to database")
	ErrTableCreationFailed = errors.New("failed to create required tables")
)