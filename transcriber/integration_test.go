package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"speakr/transcriber/internal/adapters/minio_adapter"
	"speakr/transcriber/internal/core"

	"github.com/nats-io/nats.go"
)

// Integration test to verify real adapters work with infrastructure
func TestIntegration_RecordingAndStorage(t *testing.T) {
	// Skip if not in integration test mode
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test (set INTEGRATION_TEST=true to run)")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Test MinIO storage adapter
	storage, err := minio_adapter.NewStorage(logger,
		minio_adapter.WithEndpoint("localhost:9010"),
		minio_adapter.WithCredentials("minioadmin", "minioadmin"),
		minio_adapter.WithBucket("speakr-audio"),
	)
	if err != nil {
		t.Fatalf("Failed to create MinIO storage: %v", err)
	}

	ctx := context.Background()
	recordingID := "integration-test-recording"
	mockAudioData := strings.NewReader("mock audio data for integration test")

	// Test storing audio
	filePath, err := storage.StoreAudio(ctx, recordingID, mockAudioData)
	if err != nil {
		t.Fatalf("Failed to store audio: %v", err)
	}

	t.Logf("Audio stored successfully at: %s", filePath)

	// Test retrieving audio
	reader, err := storage.RetrieveAudio(ctx, recordingID)
	if err != nil {
		t.Fatalf("Failed to retrieve audio: %v", err)
	}

	if reader == nil {
		t.Fatal("Retrieved reader is nil")
	}

	t.Log("Audio retrieved successfully")
}

func TestIntegration_NATSCommands(t *testing.T) {
	// Skip if not in integration test mode
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test (set INTEGRATION_TEST=true to run)")
	}

	// Connect to NATS
	natsConn, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		t.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer natsConn.Close()

	// Test publishing a command
	cmd := core.StartRecordingCommand{
		OutputFormat: "wav",
		Tags:         []string{"integration-test"},
		Metadata:     map[string]interface{}{"test": true},
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		t.Fatalf("Failed to marshal command: %v", err)
	}

	err = natsConn.Publish("speakr.command.recording.start", data)
	if err != nil {
		t.Fatalf("Failed to publish command: %v", err)
	}

	t.Log("Command published successfully to NATS")

	// Give some time for message to be processed
	time.Sleep(100 * time.Millisecond)
}

func TestIntegration_ErrorHandling(t *testing.T) {
	// Skip if not in integration test mode
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test (set INTEGRATION_TEST=true to run)")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Test with invalid MinIO endpoint
	_, err := minio_adapter.NewStorage(logger,
		minio_adapter.WithEndpoint("invalid-endpoint:9999"),
		minio_adapter.WithCredentials("invalid", "invalid"),
		minio_adapter.WithBucket("invalid-bucket"),
	)
	if err == nil {
		t.Error("Expected error with invalid MinIO endpoint, got nil")
	}

	logger.Info("Error handling test passed", "error", err)
}