package ports

import (
	"context"
	"io"
)

// AudioRecorder defines the interface for recording audio
type AudioRecorder interface {
	StartRecording(ctx context.Context, recordingID string, format string) error
	StopRecording(ctx context.Context, recordingID string) (io.Reader, error)
	CancelRecording(ctx context.Context, recordingID string) error
}