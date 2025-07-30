package ffmpeg_adapter

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// RecorderConfig holds configuration for the FFmpeg recorder
type RecorderConfig struct {
	TempDir       string
	InputDevice   string
	SampleRate    int
	Channels      int
	RecordTimeout time.Duration
}

// RecorderOption is a functional option for configuring the recorder
type RecorderOption func(*RecorderConfig)

// WithTempDir sets the temporary directory for audio files
func WithTempDir(dir string) RecorderOption {
	return func(c *RecorderConfig) {
		c.TempDir = dir
	}
}

// WithInputDevice sets the audio input device
func WithInputDevice(device string) RecorderOption {
	return func(c *RecorderConfig) {
		c.InputDevice = device
	}
}

// WithSampleRate sets the audio sample rate
func WithSampleRate(rate int) RecorderOption {
	return func(c *RecorderConfig) {
		c.SampleRate = rate
	}
}

// WithChannels sets the number of audio channels
func WithChannels(channels int) RecorderOption {
	return func(c *RecorderConfig) {
		c.Channels = channels
	}
}

// WithRecordTimeout sets the maximum recording duration
func WithRecordTimeout(timeout time.Duration) RecorderOption {
	return func(c *RecorderConfig) {
		c.RecordTimeout = timeout
	}
}

// Recorder implements the AudioRecorder port using FFmpeg
type Recorder struct {
	config     RecorderConfig
	logger     *slog.Logger
	recordings map[string]*recordingSession
	mu         sync.RWMutex
}

type recordingSession struct {
	cmd      *exec.Cmd
	filePath string
	cancel   context.CancelFunc
}

// NewRecorder creates a new FFmpeg recorder with functional options
func NewRecorder(logger *slog.Logger, opts ...RecorderOption) (*Recorder, error) {
	config := RecorderConfig{
		TempDir:       "/tmp/speakr",
		InputDevice:   "default",
		SampleRate:    44100,
		Channels:      1,
		RecordTimeout: 30 * time.Minute,
	}

	for _, opt := range opts {
		opt(&config)
	}

	// Ensure temp directory exists
	if err := os.MkdirAll(config.TempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return nil, ErrFFmpegNotFound
	}

	return &Recorder{
		config:     config,
		logger:     logger,
		recordings: make(map[string]*recordingSession),
	}, nil
}

// StartRecording starts a new audio recording session
func (r *Recorder) StartRecording(ctx context.Context, recordingID string, format string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	logger := r.logger.With("recording_id", recordingID, "format", format)

	// Check if recording already exists
	if _, exists := r.recordings[recordingID]; exists {
		logger.Error("Recording already in progress")
		return ErrRecordingAlreadyExists
	}

	// Create file path
	fileName := fmt.Sprintf("%s.%s", recordingID, format)
	filePath := filepath.Join(r.config.TempDir, fileName)

	// Create context with timeout
	recordCtx, cancel := context.WithTimeout(ctx, r.config.RecordTimeout)

	// Build ffmpeg command
	args := r.buildFFmpegArgs(filePath, format)
	cmd := exec.CommandContext(recordCtx, "ffmpeg", args...)

	logger.Info("Starting FFmpeg recording", "file_path", filePath, "args", args)

	// Start the recording
	if err := cmd.Start(); err != nil {
		cancel()
		logger.Error("Failed to start FFmpeg", "error", err)
		return fmt.Errorf("failed to start ffmpeg recording: %w", err)
	}

	// Store the recording session
	r.recordings[recordingID] = &recordingSession{
		cmd:      cmd,
		filePath: filePath,
		cancel:   cancel,
	}

	logger.Info("Recording started successfully")
	return nil
}

// StopRecording stops a recording and returns the audio data
func (r *Recorder) StopRecording(ctx context.Context, recordingID string) (io.Reader, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	logger := r.logger.With("recording_id", recordingID)

	session, exists := r.recordings[recordingID]
	if !exists {
		logger.Error("Recording not found")
		return nil, ErrRecordingNotFound
	}

	// Cancel the recording context to stop ffmpeg gracefully
	session.cancel()

	// Wait for the process to finish
	if err := session.cmd.Wait(); err != nil {
		// FFmpeg might exit with error when interrupted, which is expected
		logger.Warn("FFmpeg process ended with error", "error", err)
	}

	// Clean up the session
	delete(r.recordings, recordingID)

	// Check if file was created
	if _, err := os.Stat(session.filePath); os.IsNotExist(err) {
		logger.Error("Recording file not found", "file_path", session.filePath)
		return nil, ErrRecordingFileNotFound
	}

	// Open and return the file
	file, err := os.Open(session.filePath)
	if err != nil {
		logger.Error("Failed to open recording file", "error", err, "file_path", session.filePath)
		return nil, fmt.Errorf("failed to open recording file: %w", err)
	}

	logger.Info("Recording stopped successfully", "file_path", session.filePath)
	return &fileReader{file: file, filePath: session.filePath, logger: logger}, nil
}

// CancelRecording cancels an ongoing recording and cleans up
func (r *Recorder) CancelRecording(ctx context.Context, recordingID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	logger := r.logger.With("recording_id", recordingID)

	session, exists := r.recordings[recordingID]
	if !exists {
		logger.Error("Recording not found")
		return ErrRecordingNotFound
	}

	// Cancel the recording
	session.cancel()

	// Wait for process to finish
	if err := session.cmd.Wait(); err != nil {
		logger.Warn("FFmpeg process ended with error during cancellation", "error", err)
	}

	// Clean up the session
	delete(r.recordings, recordingID)

	// Remove the file if it exists
	if err := os.Remove(session.filePath); err != nil && !os.IsNotExist(err) {
		logger.Warn("Failed to remove recording file", "error", err, "file_path", session.filePath)
	}

	logger.Info("Recording cancelled successfully")
	return nil
}

// buildFFmpegArgs builds the FFmpeg command arguments
func (r *Recorder) buildFFmpegArgs(outputPath, format string) []string {
	args := []string{
		"-f", "pulse", // Use PulseAudio input (Linux)
		"-i", r.config.InputDevice,
		"-ar", fmt.Sprintf("%d", r.config.SampleRate),
		"-ac", fmt.Sprintf("%d", r.config.Channels),
		"-y", // Overwrite output file
	}

	// Add format-specific options
	switch format {
	case "wav":
		args = append(args, "-acodec", "pcm_s16le")
	case "mp3":
		args = append(args, "-acodec", "mp3", "-ab", "128k")
	default:
		// Default to WAV
		args = append(args, "-acodec", "pcm_s16le")
	}

	args = append(args, outputPath)
	return args
}

// fileReader wraps os.File and handles cleanup
type fileReader struct {
	file     *os.File
	filePath string
	logger   *slog.Logger
}

func (fr *fileReader) Read(p []byte) (n int, err error) {
	return fr.file.Read(p)
}

func (fr *fileReader) Close() error {
	err := fr.file.Close()
	// Clean up the temporary file
	if removeErr := os.Remove(fr.filePath); removeErr != nil && !os.IsNotExist(removeErr) {
		fr.logger.Warn("Failed to remove temporary file", "error", removeErr, "file_path", fr.filePath)
	}
	return err
}