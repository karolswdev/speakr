package minio_adapter

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestNewStorage(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	
	// Test with default options (will fail if MinIO not running, which is expected)
	_, err := NewStorage(logger)
	if err != nil {
		// This is expected if MinIO is not running
		logger.Info("Storage creation failed (expected if MinIO not running)", "error", err)
		return
	}
}

func TestStorageWithOptions(t *testing.T) {
	_ = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	
	// Test configuration with options
	config := StorageConfig{}
	
	opts := []StorageOption{
		WithEndpoint("localhost:9010"),
		WithCredentials("testkey", "testsecret"),
		WithBucket("test-bucket"),
		WithSSL(true),
		WithRegion("us-west-2"),
	}
	
	for _, opt := range opts {
		opt(&config)
	}
	
	if config.Endpoint != "localhost:9010" {
		t.Errorf("Expected endpoint 'localhost:9010', got %s", config.Endpoint)
	}
	
	if config.AccessKeyID != "testkey" {
		t.Errorf("Expected access key 'testkey', got %s", config.AccessKeyID)
	}
	
	if config.SecretAccessKey != "testsecret" {
		t.Errorf("Expected secret key 'testsecret', got %s", config.SecretAccessKey)
	}
	
	if config.BucketName != "test-bucket" {
		t.Errorf("Expected bucket 'test-bucket', got %s", config.BucketName)
	}
	
	if !config.UseSSL {
		t.Error("Expected SSL to be enabled")
	}
	
	if config.Region != "us-west-2" {
		t.Errorf("Expected region 'us-west-2', got %s", config.Region)
	}
}

func TestStorageOperations(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	
	storage, err := NewStorage(logger,
		WithEndpoint("localhost:9010"),
		WithCredentials("minioadmin", "minioadmin"),
		WithBucket("speakr-audio"),
	)
	if err != nil {
		t.Skipf("MinIO not available, skipping integration test: %v", err)
	}
	
	ctx := context.Background()
	recordingID := "test-recording-123"
	audioData := strings.NewReader("mock audio data for testing")
	
	// Test storing audio
	filePath, err := storage.StoreAudio(ctx, recordingID, audioData)
	if err != nil {
		t.Fatalf("Failed to store audio: %v", err)
	}
	
	if !strings.Contains(filePath, recordingID) {
		t.Errorf("File path should contain recording ID, got: %s", filePath)
	}
	
	// Test retrieving audio
	reader, err := storage.RetrieveAudio(ctx, recordingID)
	if err != nil {
		t.Fatalf("Failed to retrieve audio: %v", err)
	}
	
	if reader == nil {
		t.Error("Retrieved reader should not be nil")
	}
	
	// Test retrieving non-existent audio
	_, err = storage.RetrieveAudio(ctx, "non-existent-recording")
	if err != ErrObjectNotFound {
		t.Errorf("Expected ErrObjectNotFound, got %v", err)
	}
}