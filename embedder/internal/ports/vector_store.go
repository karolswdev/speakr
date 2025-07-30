package ports

import (
	"context"
)

// VectorRecord represents a complete record to be stored in the vector database
type VectorRecord struct {
	RecordingID      string    `json:"recording_id"`
	TranscribedText  string    `json:"transcribed_text"`
	Tags             []string  `json:"tags"`
	Embedding        []float32 `json:"embedding"`
}

// VectorStore defines the interface for storing vector embeddings and associated data
type VectorStore interface {
	StoreRecord(ctx context.Context, record VectorRecord) error
	GetRecord(ctx context.Context, recordingID string) (*VectorRecord, error)
}