package ffmpeg_adapter

import (
	"context"
	"runtime"
	"testing"
	"time"
)

func TestNewDeviceDetector(t *testing.T) {
	detector := NewDeviceDetector()
	
	if detector == nil {
		t.Fatal("Expected device detector to be created, got nil")
	}
	
	if detector.platform != runtime.GOOS {
		t.Errorf("Expected platform %s, got %s", runtime.GOOS, detector.platform)
	}
}

func TestGetAudioSubsystem(t *testing.T) {
	detector := NewDeviceDetector()
	
	subsystem := detector.GetAudioSubsystem()
	
	// Should return a valid subsystem for any platform
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

func TestValidateInputDevice_Default(t *testing.T) {
	detector := NewDeviceDetector()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Default device should always be valid
	err := detector.ValidateInputDevice(ctx, "default")
	if err != nil {
		t.Errorf("Expected default device to be valid, got error: %v", err)
	}
}

func TestValidateInputDevice_NonExistent(t *testing.T) {
	detector := NewDeviceDetector()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Non-existent device should return error
	err := detector.ValidateInputDevice(ctx, "non-existent-device-12345")
	if err != ErrDeviceNotFound {
		t.Errorf("Expected ErrDeviceNotFound, got: %v", err)
	}
}

func TestValidateOutputDevice_Default(t *testing.T) {
	detector := NewDeviceDetector()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Default device should always be valid
	err := detector.ValidateOutputDevice(ctx, "default")
	if err != nil {
		t.Errorf("Expected default device to be valid, got error: %v", err)
	}
}

func TestListInputDevices(t *testing.T) {
	detector := NewDeviceDetector()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	devices, err := detector.ListInputDevices(ctx)
	if err != nil {
		t.Logf("Warning: Could not list input devices: %v", err)
		return // Skip test if device listing fails (common in CI environments)
	}
	
	// Should have at least one device (default)
	if len(devices) == 0 {
		t.Error("Expected at least one input device")
	}
	
	// Check device structure
	for _, device := range devices {
		if device.ID == "" {
			t.Error("Device ID should not be empty")
		}
		if device.Name == "" {
			t.Error("Device name should not be empty")
		}
		if device.Type != "input" {
			t.Errorf("Expected device type 'input', got: %s", device.Type)
		}
	}
}

func TestListOutputDevices(t *testing.T) {
	detector := NewDeviceDetector()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	devices, err := detector.ListOutputDevices(ctx)
	if err != nil {
		t.Logf("Warning: Could not list output devices: %v", err)
		return // Skip test if device listing fails (common in CI environments)
	}
	
	// Should have at least one device (default)
	if len(devices) == 0 {
		t.Error("Expected at least one output device")
	}
	
	// Check device structure
	for _, device := range devices {
		if device.ID == "" {
			t.Error("Device ID should not be empty")
		}
		if device.Name == "" {
			t.Error("Device name should not be empty")
		}
		if device.Type != "output" {
			t.Errorf("Expected device type 'output', got: %s", device.Type)
		}
	}
}

func TestParseMacOSDevices(t *testing.T) {
	detector := NewDeviceDetector()
	
	// Mock ffmpeg output for macOS
	mockOutput := `
ffmpeg version 4.4.0 Copyright (c) 2000-2021 the FFmpeg developers
[AVFoundation input device @ 0x7f8b1c004000] AVFoundation video devices:
[AVFoundation input device @ 0x7f8b1c004000] [0] FaceTime HD Camera
[AVFoundation input device @ 0x7f8b1c004000] AVFoundation audio devices:
[AVFoundation input device @ 0x7f8b1c004000] [0] Built-in Microphone
[AVFoundation input device @ 0x7f8b1c004000] [1] External Microphone
`
	
	devices := detector.parseMacOSDevices(mockOutput, "input")
	
	if len(devices) != 2 {
		t.Errorf("Expected 2 devices, got %d", len(devices))
	}
	
	if devices[0].ID != "0" || devices[0].Name != "Built-in Microphone" {
		t.Errorf("Unexpected first device: %+v", devices[0])
	}
	
	if devices[1].ID != "1" || devices[1].Name != "External Microphone" {
		t.Errorf("Unexpected second device: %+v", devices[1])
	}
}

func TestParseWindowsDevices(t *testing.T) {
	detector := NewDeviceDetector()
	
	// Mock ffmpeg output for Windows
	mockOutput := `
ffmpeg version 4.4.0 Copyright (c) 2000-2021 the FFmpeg developers
[dshow @ 0x000001a2b4004000] DirectShow video devices (some may be both video and audio devices)
[dshow @ 0x000001a2b4004000] DirectShow audio devices
[dshow @ 0x000001a2b4004000]  "Microphone (Realtek Audio)"
[dshow @ 0x000001a2b4004000]  "Line In (Realtek Audio)"
`
	
	devices := detector.parseWindowsDevices(mockOutput, "input")
	
	if len(devices) != 2 {
		t.Errorf("Expected 2 devices, got %d", len(devices))
	}
	
	if devices[0].Name != "Microphone (Realtek Audio)" {
		t.Errorf("Unexpected first device: %+v", devices[0])
	}
	
	if devices[1].Name != "Line In (Realtek Audio)" {
		t.Errorf("Unexpected second device: %+v", devices[1])
	}
}

// Integration test - only run when INTEGRATION_TEST=true
func TestDeviceDetection_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	detector := NewDeviceDetector()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	t.Run("ListInputDevices", func(t *testing.T) {
		devices, err := detector.ListInputDevices(ctx)
		if err != nil {
			t.Logf("Could not list input devices (expected in CI): %v", err)
			return
		}
		
		t.Logf("Found %d input devices:", len(devices))
		for _, device := range devices {
			t.Logf("  - %s: %s (default: %v, available: %v)", 
				device.ID, device.Name, device.IsDefault, device.IsAvailable)
		}
	})
	
	t.Run("ListOutputDevices", func(t *testing.T) {
		devices, err := detector.ListOutputDevices(ctx)
		if err != nil {
			t.Logf("Could not list output devices (expected in CI): %v", err)
			return
		}
		
		t.Logf("Found %d output devices:", len(devices))
		for _, device := range devices {
			t.Logf("  - %s: %s (default: %v, available: %v)", 
				device.ID, device.Name, device.IsDefault, device.IsAvailable)
		}
	})
	
	t.Run("AudioSubsystem", func(t *testing.T) {
		subsystem := detector.GetAudioSubsystem()
		t.Logf("Detected audio subsystem: %s", subsystem)
	})
}