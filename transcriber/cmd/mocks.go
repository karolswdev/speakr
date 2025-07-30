package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"
)

// Mock implementations for development and testing

type mockAudioRecorder struct {
	logger *slog.Logger
}

func (m *mockAudioRecorder) StartRecording(ctx context.Context, recordingID string, format string) error {
	m.logger.Info("Mock: Starting recording", "recording_id", recordingID, "format", format)
	return nil
}

func (m *mockAudioRecorder) StopRecording(ctx context.Context, recordingID string) (io.Reader, error) {
	m.logger.Info("Mock: Stopping recording", "recording_id", recordingID)
	// Return mock audio data
	mockAudioData := "mock audio data for " + recordingID
	return strings.NewReader(mockAudioData), nil
}

func (m *mockAudioRecorder) CancelRecording(ctx context.Context, recordingID string) error {
	m.logger.Info("Mock: Cancelling recording", "recording_id", recordingID)
	return nil
}

type mockTranscriptionService struct {
	logger *slog.Logger
}

func (m *mockTranscriptionService) TranscribeAudio(ctx context.Context, audioData io.Reader, format string) (string, error) {
	m.logger.Info("Mock: Transcribing audio", "format", format)
	// Simulate some processing time
	time.Sleep(100 * time.Millisecond)
	return "This is a mock transcription of the audio content.", nil
}

type mockObjectStore struct {
	logger *slog.Logger
	store  map[string]string
}

func (m *mockObjectStore) StoreAudio(ctx context.Context, recordingID string, audioData io.Reader) (string, error) {
	m.logger.Info("Mock: Storing audio", "recording_id", recordingID)
	
	if m.store == nil {
		m.store = make(map[string]string)
	}
	
	// Read the audio data
	data, err := io.ReadAll(audioData)
	if err != nil {
		return "", fmt.Errorf("failed to read audio data: %w", err)
	}
	
	// Store in mock storage
	m.store[recordingID] = string(data)
	
	filePath := fmt.Sprintf("/mock/storage/%s.wav", recordingID)
	return filePath, nil
}

func (m *mockObjectStore) RetrieveAudio(ctx context.Context, recordingID string) (io.Reader, error) {
	m.logger.Info("Mock: Retrieving audio", "recording_id", recordingID)
	
	if m.store == nil {
		return nil, fmt.Errorf("audio file not found: %s", recordingID)
	}
	
	data, exists := m.store[recordingID]
	if !exists {
		return nil, fmt.Errorf("audio file not found: %s", recordingID)
	}
	
	return strings.NewReader(data), nil
}