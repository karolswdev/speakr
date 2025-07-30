package ffmpeg_adapter

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestNewRecorder(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	
	// Test with default options
	recorder, err := NewRecorder(logger)
	if err != nil {
		// Skip test if ffmpeg is not available
		if err == ErrFFmpegNotFound {
			t.Skip("FFmpeg not found, skipping test")
		}
		t.Fatalf("Failed to create recorder: %v", err)
	}
	
	if recorder.config.TempDir != "/tmp/speakr" {
		t.Errorf("Expected temp dir '/tmp/speakr', got %s", recorder.config.TempDir)
	}
	
	if recorder.config.SampleRate != 44100 {
		t.Errorf("Expected sample rate 44100, got %d", recorder.config.SampleRate)
	}
}

func TestRecorderWithOptions(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	
	recorder, err := NewRecorder(logger,
		WithTempDir("/tmp/test"),
		WithSampleRate(48000),
		WithChannels(2),
		WithRecordTimeout(5*time.Minute),
	)
	if err != nil {
		if err == ErrFFmpegNotFound {
			t.Skip("FFmpeg not found, skipping test")
		}
		t.Fatalf("Failed to create recorder: %v", err)
	}
	
	if recorder.config.TempDir != "/tmp/test" {
		t.Errorf("Expected temp dir '/tmp/test', got %s", recorder.config.TempDir)
	}
	
	if recorder.config.SampleRate != 48000 {
		t.Errorf("Expected sample rate 48000, got %d", recorder.config.SampleRate)
	}
	
	if recorder.config.Channels != 2 {
		t.Errorf("Expected channels 2, got %d", recorder.config.Channels)
	}
	
	if recorder.config.RecordTimeout != 5*time.Minute {
		t.Errorf("Expected timeout 5m, got %v", recorder.config.RecordTimeout)
	}
}

func TestBuildFFmpegArgs(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	
	recorder, err := NewRecorder(logger)
	if err != nil {
		if err == ErrFFmpegNotFound {
			t.Skip("FFmpeg not found, skipping test")
		}
		t.Fatalf("Failed to create recorder: %v", err)
	}
	
	args := recorder.buildFFmpegArgs("/tmp/test.wav", "wav")
	
	// Check that essential arguments are present
	expectedArgs := []string{"-f", "pulse", "-i", "default", "-ar", "44100", "-ac", "1", "-y", "-acodec", "pcm_s16le", "/tmp/test.wav"}
	
	if len(args) != len(expectedArgs) {
		t.Errorf("Expected %d args, got %d", len(expectedArgs), len(args))
	}
	
	for i, expected := range expectedArgs {
		if i < len(args) && args[i] != expected {
			t.Errorf("Expected arg[%d] = %s, got %s", i, expected, args[i])
		}
	}
}

func TestRecordingErrors(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	
	recorder, err := NewRecorder(logger)
	if err != nil {
		if err == ErrFFmpegNotFound {
			t.Skip("FFmpeg not found, skipping test")
		}
		t.Fatalf("Failed to create recorder: %v", err)
	}
	
	ctx := context.Background()
	
	// Test stopping non-existent recording
	_, err = recorder.StopRecording(ctx, "non-existent")
	if err != ErrRecordingNotFound {
		t.Errorf("Expected ErrRecordingNotFound, got %v", err)
	}
	
	// Test cancelling non-existent recording
	err = recorder.CancelRecording(ctx, "non-existent")
	if err != ErrRecordingNotFound {
		t.Errorf("Expected ErrRecordingNotFound, got %v", err)
	}
}