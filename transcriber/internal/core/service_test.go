package core

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"speakr/transcriber/internal/ports"
)

// Mock implementations for testing
type mockAudioRecorder struct {
	startRecordingFunc  func(ctx context.Context, recordingID string, format string) error
	stopRecordingFunc   func(ctx context.Context, recordingID string) (io.Reader, error)
	cancelRecordingFunc func(ctx context.Context, recordingID string) error
}

func (m *mockAudioRecorder) StartRecording(ctx context.Context, recordingID string, format string) error {
	if m.startRecordingFunc != nil {
		return m.startRecordingFunc(ctx, recordingID, format)
	}
	return nil
}

func (m *mockAudioRecorder) StopRecording(ctx context.Context, recordingID string) (io.Reader, error) {
	if m.stopRecordingFunc != nil {
		return m.stopRecordingFunc(ctx, recordingID)
	}
	return strings.NewReader("mock audio data"), nil
}

func (m *mockAudioRecorder) CancelRecording(ctx context.Context, recordingID string) error {
	if m.cancelRecordingFunc != nil {
		return m.cancelRecordingFunc(ctx, recordingID)
	}
	return nil
}

type mockTranscriptionService struct {
	transcribeAudioFunc func(ctx context.Context, audioData io.Reader, format string) (string, error)
}

func (m *mockTranscriptionService) TranscribeAudio(ctx context.Context, audioData io.Reader, format string) (string, error) {
	if m.transcribeAudioFunc != nil {
		return m.transcribeAudioFunc(ctx, audioData, format)
	}
	return "mock transcription", nil
}

type mockObjectStore struct {
	storeAudioFunc    func(ctx context.Context, recordingID string, audioData io.Reader) (string, error)
	retrieveAudioFunc func(ctx context.Context, recordingID string) (io.Reader, error)
}

func (m *mockObjectStore) StoreAudio(ctx context.Context, recordingID string, audioData io.Reader) (string, error) {
	if m.storeAudioFunc != nil {
		return m.storeAudioFunc(ctx, recordingID, audioData)
	}
	return "/mock/path/" + recordingID + ".wav", nil
}

func (m *mockObjectStore) RetrieveAudio(ctx context.Context, recordingID string) (io.Reader, error) {
	if m.retrieveAudioFunc != nil {
		return m.retrieveAudioFunc(ctx, recordingID)
	}
	return strings.NewReader("mock audio data"), nil
}

type mockEventPublisher struct {
	publishEventFunc func(ctx context.Context, event ports.Event) error
	publishedEvents  []ports.Event
}

func (m *mockEventPublisher) PublishEvent(ctx context.Context, event ports.Event) error {
	m.publishedEvents = append(m.publishedEvents, event)
	if m.publishEventFunc != nil {
		return m.publishEventFunc(ctx, event)
	}
	return nil
}

func createTestService() (*Service, *mockAudioRecorder, *mockTranscriptionService, *mockObjectStore, *mockEventPublisher) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	
	audioRecorder := &mockAudioRecorder{}
	transcriptionSvc := &mockTranscriptionService{}
	objectStore := &mockObjectStore{}
	eventPublisher := &mockEventPublisher{}
	
	service := NewService(audioRecorder, transcriptionSvc, objectStore, eventPublisher, logger)
	
	return service, audioRecorder, transcriptionSvc, objectStore, eventPublisher
}

func TestService_StartRecording(t *testing.T) {
	service, _, _, _, eventPublisher := createTestService()
	
	ctx := context.Background()
	cmd := StartRecordingCommand{
		OutputFormat: "wav",
		Tags:         []string{"test", "recording"},
		Metadata:     map[string]interface{}{"source": "test"},
	}
	
	// Test successful start recording
	err := service.StartRecording(ctx, cmd)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Verify event was published
	if len(eventPublisher.publishedEvents) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(eventPublisher.publishedEvents))
	}
	
	event := eventPublisher.publishedEvents[0]
	if event.Subject != "speakr.event.recording.started" {
		t.Errorf("Expected subject 'speakr.event.recording.started', got %s", event.Subject)
	}
}

func TestService_StopRecording(t *testing.T) {
	service, _, _, _, eventPublisher := createTestService()
	
	ctx := context.Background()
	cmd := StopRecordingCommand{
		RecordingID:      "test-recording-id",
		TranscribeOnStop: false,
		Metadata:         map[string]interface{}{"source": "test"},
	}
	
	// Test successful stop recording
	err := service.StopRecording(ctx, cmd)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Verify event was published
	if len(eventPublisher.publishedEvents) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(eventPublisher.publishedEvents))
	}
	
	event := eventPublisher.publishedEvents[0]
	if event.Subject != "speakr.event.recording.finished" {
		t.Errorf("Expected subject 'speakr.event.recording.finished', got %s", event.Subject)
	}
}

func TestService_CancelRecording(t *testing.T) {
	service, _, _, _, eventPublisher := createTestService()
	
	ctx := context.Background()
	cmd := CancelRecordingCommand{
		RecordingID: "test-recording-id",
	}
	
	// Test successful cancel recording
	err := service.CancelRecording(ctx, cmd)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Verify event was published
	if len(eventPublisher.publishedEvents) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(eventPublisher.publishedEvents))
	}
	
	event := eventPublisher.publishedEvents[0]
	if event.Subject != "speakr.event.recording.cancelled" {
		t.Errorf("Expected subject 'speakr.event.recording.cancelled', got %s", event.Subject)
	}
}

func TestService_TranscribeAudio(t *testing.T) {
	service, _, _, _, eventPublisher := createTestService()
	
	ctx := context.Background()
	cmd := TranscriptionCommand{
		RecordingID: "test-recording-id",
		Tags:        []string{"test", "transcription"},
		Metadata:    map[string]interface{}{"source": "test"},
	}
	
	// Test successful transcription
	err := service.TranscribeAudio(ctx, cmd)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Verify event was published
	if len(eventPublisher.publishedEvents) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(eventPublisher.publishedEvents))
	}
	
	event := eventPublisher.publishedEvents[0]
	if event.Subject != "speakr.event.transcription.succeeded" {
		t.Errorf("Expected subject 'speakr.event.transcription.succeeded', got %s", event.Subject)
	}
}