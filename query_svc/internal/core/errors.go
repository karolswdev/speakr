package core

import "errors"

var (
	// ErrInvalidQuery indicates the query text is invalid or empty
	ErrInvalidQuery = errors.New("invalid query text")
	
	// ErrEmbeddingFailed indicates the embedding generation failed
	ErrEmbeddingFailed = errors.New("embedding generation failed")
	
	// ErrSearchFailed indicates the vector search failed
	ErrSearchFailed = errors.New("vector search failed")
	
	// ErrDatabaseUnavailable indicates the database is not accessible
	ErrDatabaseUnavailable = errors.New("database unavailable")
)