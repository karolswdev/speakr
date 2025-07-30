package ports

import (
	"context"
	"io"
)

// TranscriptionService defines the interface for transcribing audio
type TranscriptionService interface {
	TranscribeAudio(ctx context.Context, audioData io.Reader, format string) (string, error)
}