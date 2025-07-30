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

	// ErrDeviceNotFound indicates that the specified audio device was not found
	ErrDeviceNotFound = errors.New("audio device not found")

	// ErrDevicePermissionDenied indicates permission issues accessing the audio device
	ErrDevicePermissionDenied = errors.New("permission denied accessing audio device")

	// ErrDeviceBusy indicates that the audio device is currently in use
	ErrDeviceBusy = errors.New("audio device is busy")

	// ErrUnsupportedPlatform indicates that the current platform is not supported
	ErrUnsupportedPlatform = errors.New("unsupported platform for audio device detection")
)