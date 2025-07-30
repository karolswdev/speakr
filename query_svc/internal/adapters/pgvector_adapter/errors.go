package pgvector_adapter

import "errors"

var (
	// ErrConnectionFailed indicates database connection failed
	ErrConnectionFailed = errors.New("database connection failed")
	
	// ErrQueryFailed indicates the search query failed
	ErrQueryFailed = errors.New("search query failed")
	
	// ErrInvalidEmbedding indicates the embedding vector is invalid
	ErrInvalidEmbedding = errors.New("invalid embedding vector")
)