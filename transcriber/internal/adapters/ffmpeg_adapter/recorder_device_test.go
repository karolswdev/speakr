package ffmpeg_adapter

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestRecorder_WithDeviceConfiguration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	recorder, err := NewRecorder(logger,
		WithTempDir("/tmp/test-speakr"),
		WithInputDevice("test-input"),
		WithOutputDevice("test-output"),
		WithSampleRate(22050),
		WithChannels(2),
	)
	
	if err != nil {
		t.Fatalf("Failed to create recorder: %v", err)
	}
	
	if recorder.config.InputDevice != "test-input" {
		t.Errorf("Expected input device 'test-input', got: %s", recorder.config.InputDevice)
	}
	
	if recorder.config.OutputDevice != "test-output" {
		t.Errorf("Expected output device 'test-output', got: %s", recorder.config.OutputDevice)
	}
	
	if recorder.config.SampleRate != 22050 {
		t.Errorf("Expected sample rate 22050, got: %d", recorder.config.SampleRate)
	}
	
	if recorder.config.Channels != 2 {
		t.Errorf("Expected 2 channels, got: %d", recorder.config.Channels)
	}
}

func TestRecorder_DeviceValidation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	// Test with non-existent device
	recorder, err := NewRecorder(logger,
		WithInputDevice("non-existent-device-12345"),
	)
	
	if err != nil {
		t.Fatalf("Failed to create recorder: %v", err)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// This should fail device validation
	err = recorder.StartRecording(ctx, "test-recording", "wav")
	if err == nil {
		t.Error("Expected device validation to fail for non-existent device")
		// Clean up if recording somehow started
		recorder.CancelRecording(ctx, "test-recording")
	}
}

func TestRecorder_DefaultDeviceValidation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	// Test with default device (should not validate)
	recorder, err := NewRecorder(logger,
		WithInputDevice("default"),
	)
	
	if err != nil {
		t.Fatalf("Failed to create recorder: %v", err)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// This should not perform device validation for "default"
	err = recorder.StartRecording(ctx, "test-recording", "wav")
	if err != nil && err != ErrFFmpegNotFound {
		t.Errorf("Unexpected error for default device: %v", err)
	}
	
	// Clean up if recording started
	if err == nil {
		recorder.CancelRecording(ctx, "test-recording")
	}
}

func TestRecorder_BuildFFmpegArgs_CrossPlatform(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	recorder, err := NewRecorder(logger,
		WithInputDevice("test-device"),
		WithSampleRate(44100),
		WithChannels(1),
	)
	
	if err != nil {
		t.Fatalf("Failed to create recorder: %v", err)
	}
	
	args := recorder.buildFFmpegArgs("/tmp/test.wav", "wav")
	
	// Check that args contain expected elements
	expectedElements := []string{"-f", "test-device", "-ar", "44100", "-ac", "1", "-y"}
	
	for _, expected := range expectedElements {
		found := false
		for _, arg := range args {
			if arg == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected argument '%s' not found in: %v", expected, args)
		}
	}
	
	// Check that audio subsystem is set (not hardcoded to "pulse")
	if len(args) >= 2 && args[0] == "-f" {
		subsystem := args[1]
		validSubsystems := []string{"pulse", "alsa", "avfoundation", "dshow"}
		found := false
		for _, valid := range validSubsystems {
			if subsystem == valid {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected valid audio subsystem, got: %s", subsystem)
		}
	}
}

// Integration test for device enumeration
func TestRecorder_DeviceEnumeration_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	recorder, err := NewRecorder(logger)
	if err != nil {
		t.Fatalf("Failed to create recorder: %v", err)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	t.Run("ListInputDevices", func(t *testing.T) {
		devices, err := recorder.deviceDetector.ListInputDevices(ctx)
		if err != nil {
			t.Logf("Could not list input devices (expected in CI): %v", err)
			return
		}
		
		t.Logf("Available input devices:")
		for _, device := range devices {
			t.Logf("  - ID: %s, Name: %s, Default: %v", 
				device.ID, device.Name, device.IsDefault)
		}
	})
	
	t.Run("ListOutputDevices", func(t *testing.T) {
		devices, err := recorder.deviceDetector.ListOutputDevices(ctx)
		if err != nil {
			t.Logf("Could not list output devices (expected in CI): %v", err)
			return
		}
		
		t.Logf("Available output devices:")
		for _, device := range devices {
			t.Logf("  - ID: %s, Name: %s, Default: %v", 
				device.ID, device.Name, device.IsDefault)
		}
	})
}