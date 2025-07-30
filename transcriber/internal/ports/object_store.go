package ports

import (
	"context"
	"io"
)

// ObjectStore defines the interface for storing and retrieving audio files
type ObjectStore interface {
	StoreAudio(ctx context.Context, recordingID string, audioData io.Reader) (string, error)
	RetrieveAudio(ctx context.Context, recordingID string) (io.Reader, error)
}