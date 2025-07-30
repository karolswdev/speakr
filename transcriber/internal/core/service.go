package core

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"speakr/transcriber/internal/ports"

	"github.com/google/uuid"
)

// Service represents the core transcriber service
type Service struct {
	audioRecorder       ports.AudioRecorder
	transcriptionSvc    ports.TranscriptionService
	objectStore         ports.ObjectStore
	eventPublisher      ports.EventPublisher
	logger              *slog.Logger
}

// NewService creates a new transcriber service
func NewService(
	audioRecorder ports.AudioRecorder,
	transcriptionSvc ports.TranscriptionService,
	objectStore ports.ObjectStore,
	eventPublisher ports.EventPublisher,
	logger *slog.Logger,
) *Service {
	return &Service{
		audioRecorder:    audioRecorder,
		transcriptionSvc: transcriptionSvc,
		objectStore:      objectStore,
		eventPublisher:   eventPublisher,
		logger:           logger,
	}
}

// StartRecordingCommand represents the start recording command payload
type StartRecordingCommand struct {
	OutputFormat string                 `json:"output_format"`
	Tags         []string               `json:"tags"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// StopRecordingCommand represents the stop recording command payload
type StopRecordingCommand struct {
	RecordingID      string                 `json:"recording_id"`
	TranscribeOnStop bool                   `json:"transcribe_on_stop"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// CancelRecordingCommand represents the cancel recording command payload
type CancelRecordingCommand struct {
	RecordingID string `json:"recording_id"`
}

// TranscriptionCommand represents the transcription command payload
type TranscriptionCommand struct {
	RecordingID string                 `json:"recording_id,omitempty"`
	AudioData   string                 `json:"audio_data,omitempty"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// StartRecording handles the start recording command
func (s *Service) StartRecording(ctx context.Context, cmd StartRecordingCommand) error {
	recordingID := uuid.New().String()
	
	correlationID := s.getCorrelationID(ctx)
	logger := s.logger.With(
		"correlation_id", correlationID,
		"recording_id", recordingID,
		"operation", "start_recording",
	)

	logger.Info("Starting recording", "format", cmd.OutputFormat, "tags", cmd.Tags)

	err := s.audioRecorder.StartRecording(ctx, recordingID, cmd.OutputFormat)
	if err != nil {
		logger.Error("Failed to start recording", "error", err)
		return fmt.Errorf("failed to start recording: %w", err)
	}

	// Publish recording started event
	event := ports.Event{
		Subject: "speakr.event.recording.started",
		Data: map[string]interface{}{
			"recording_id": recordingID,
			"tags":         cmd.Tags,
			"metadata":     cmd.Metadata,
		},
	}

	if err := s.eventPublisher.PublishEvent(ctx, event); err != nil {
		logger.Error("Failed to publish recording started event", "error", err)
		return fmt.Errorf("failed to publish recording started event: %w", err)
	}

	logger.Info("Recording started successfully")
	return nil
}

// StopRecording handles the stop recording command
func (s *Service) StopRecording(ctx context.Context, cmd StopRecordingCommand) error {
	correlationID := s.getCorrelationID(ctx)
	logger := s.logger.With(
		"correlation_id", correlationID,
		"recording_id", cmd.RecordingID,
		"operation", "stop_recording",
	)

	logger.Info("Stopping recording", "transcribe_on_stop", cmd.TranscribeOnStop)

	audioData, err := s.audioRecorder.StopRecording(ctx, cmd.RecordingID)
	if err != nil {
		logger.Error("Failed to stop recording", "error", err)
		return fmt.Errorf("failed to stop recording: %w", err)
	}

	// Store audio file
	audioFilePath, err := s.objectStore.StoreAudio(ctx, cmd.RecordingID, audioData)
	if err != nil {
		logger.Error("Failed to store audio file", "error", err)
		return fmt.Errorf("failed to store audio file: %w", err)
	}

	// Publish recording finished event
	event := ports.Event{
		Subject: "speakr.event.recording.finished",
		Data: map[string]interface{}{
			"recording_id":    cmd.RecordingID,
			"audio_file_path": audioFilePath,
			"tags":            []string{}, // Will be populated from original command
			"metadata":        cmd.Metadata,
		},
	}

	if err := s.eventPublisher.PublishEvent(ctx, event); err != nil {
		logger.Error("Failed to publish recording finished event", "error", err)
		return fmt.Errorf("failed to publish recording finished event: %w", err)
	}

	// If transcribe_on_stop is true, trigger transcription
	if cmd.TranscribeOnStop {
		transcribeCmd := TranscriptionCommand{
			RecordingID: cmd.RecordingID,
			Tags:        []string{}, // Will be populated from original command
			Metadata:    cmd.Metadata,
		}
		if err := s.TranscribeAudio(ctx, transcribeCmd); err != nil {
			logger.Error("Failed to transcribe audio after stop", "error", err)
			return fmt.Errorf("failed to transcribe audio after stop: %w", err)
		}
	}

	logger.Info("Recording stopped successfully", "audio_file_path", audioFilePath)
	return nil
}

// CancelRecording handles the cancel recording command
func (s *Service) CancelRecording(ctx context.Context, cmd CancelRecordingCommand) error {
	correlationID := s.getCorrelationID(ctx)
	logger := s.logger.With(
		"correlation_id", correlationID,
		"recording_id", cmd.RecordingID,
		"operation", "cancel_recording",
	)

	logger.Info("Cancelling recording")

	err := s.audioRecorder.CancelRecording(ctx, cmd.RecordingID)
	if err != nil {
		logger.Error("Failed to cancel recording", "error", err)
		return fmt.Errorf("failed to cancel recording: %w", err)
	}

	// Publish recording cancelled event
	event := ports.Event{
		Subject: "speakr.event.recording.cancelled",
		Data: map[string]interface{}{
			"recording_id": cmd.RecordingID,
		},
	}

	if err := s.eventPublisher.PublishEvent(ctx, event); err != nil {
		logger.Error("Failed to publish recording cancelled event", "error", err)
		return fmt.Errorf("failed to publish recording cancelled event: %w", err)
	}

	logger.Info("Recording cancelled successfully")
	return nil
}

// TranscribeAudio handles the transcription command
func (s *Service) TranscribeAudio(ctx context.Context, cmd TranscriptionCommand) error {
	correlationID := s.getCorrelationID(ctx)
	logger := s.logger.With(
		"correlation_id", correlationID,
		"recording_id", cmd.RecordingID,
		"operation", "transcribe_audio",
	)

	logger.Info("Starting transcription")

	var audioData io.Reader
	var err error

	if cmd.RecordingID != "" {
		// Retrieve audio from object store
		audioData, err = s.objectStore.RetrieveAudio(ctx, cmd.RecordingID)
		if err != nil {
			logger.Error("Failed to retrieve audio file", "error", err)
			
			// Publish transcription failed event
			failEvent := ports.Event{
				Subject: "speakr.event.transcription.failed",
				Data: map[string]interface{}{
					"recording_id": cmd.RecordingID,
					"error":        "Failed to retrieve audio file",
					"tags":         cmd.Tags,
					"metadata":     cmd.Metadata,
				},
			}
			s.eventPublisher.PublishEvent(ctx, failEvent)
			
			return fmt.Errorf("failed to retrieve audio file: %w", err)
		}
	} else if cmd.AudioData != "" {
		// Handle base64 encoded audio data
		// This would need proper base64 decoding implementation
		logger.Error("Base64 audio data not yet implemented")
		return fmt.Errorf("base64 audio data not yet implemented")
	} else {
		logger.Error("Neither recording_id nor audio_data provided")
		return fmt.Errorf("neither recording_id nor audio_data provided")
	}

	// Transcribe the audio
	transcribedText, err := s.transcriptionSvc.TranscribeAudio(ctx, audioData, "wav")
	if err != nil {
		logger.Error("Failed to transcribe audio", "error", err)
		
		// Publish transcription failed event
		failEvent := ports.Event{
			Subject: "speakr.event.transcription.failed",
			Data: map[string]interface{}{
				"recording_id": cmd.RecordingID,
				"error":        err.Error(),
				"tags":         cmd.Tags,
				"metadata":     cmd.Metadata,
			},
		}
		s.eventPublisher.PublishEvent(ctx, failEvent)
		
		return fmt.Errorf("failed to transcribe audio: %w", err)
	}

	// Publish transcription succeeded event
	event := ports.Event{
		Subject: "speakr.event.transcription.succeeded",
		Data: map[string]interface{}{
			"recording_id":     cmd.RecordingID,
			"transcribed_text": transcribedText,
			"tags":             cmd.Tags,
			"metadata":         cmd.Metadata,
		},
	}

	if err := s.eventPublisher.PublishEvent(ctx, event); err != nil {
		logger.Error("Failed to publish transcription succeeded event", "error", err)
		return fmt.Errorf("failed to publish transcription succeeded event: %w", err)
	}

	logger.Info("Transcription completed successfully", "text_length", len(transcribedText))
	return nil
}

func (s *Service) getCorrelationID(ctx context.Context) string {
	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		if id, ok := correlationID.(string); ok {
			return id
		}
	}
	return uuid.New().String()
}