package ffmpeg_adapter

import "errors"

// Custom error types for predictable failures
var (
	ErrFFmpegNotFound           = errors.New("ffmpeg executable not found in PATH")
	ErrRecordingAlreadyExists   = errors.New("recording with this ID already exists")
	ErrRecordingNotFound        = errors.New("recording with this ID not found")
	ErrRecordingFileNotFound    = errors.New("recording file not found after stopping")
	ErrInvalidFormat            = errors.New("invalid audio format specified")
	ErrInsufficientDiskSpace    = errors.New("insufficient disk space for recording")
	ErrPermissionDenied         = errors.New("permission denied accessing audio device or file system")
)