package ffmpeg_adapter

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// AudioDevice represents an audio device
type AudioDevice struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"` // "input" or "output"
	IsDefault   bool   `json:"is_default"`
	IsAvailable bool   `json:"is_available"`
}

// DeviceDetector handles audio device detection and validation
type DeviceDetector struct {
	platform string
}

// NewDeviceDetector creates a new device detector
func NewDeviceDetector() *DeviceDetector {
	return &DeviceDetector{
		platform: runtime.GOOS,
	}
}

// ListInputDevices returns a list of available audio input devices
func (d *DeviceDetector) ListInputDevices(ctx context.Context) ([]AudioDevice, error) {
	switch d.platform {
	case "linux":
		return d.listLinuxInputDevices(ctx)
	case "darwin":
		return d.listMacOSInputDevices(ctx)
	case "windows":
		return d.listWindowsInputDevices(ctx)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", d.platform)
	}
}

// ListOutputDevices returns a list of available audio output devices
func (d *DeviceDetector) ListOutputDevices(ctx context.Context) ([]AudioDevice, error) {
	switch d.platform {
	case "linux":
		return d.listLinuxOutputDevices(ctx)
	case "darwin":
		return d.listMacOSOutputDevices(ctx)
	case "windows":
		return d.listWindowsOutputDevices(ctx)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", d.platform)
	}
}

// ValidateInputDevice checks if an input device is available
func (d *DeviceDetector) ValidateInputDevice(ctx context.Context, deviceID string) error {
	devices, err := d.ListInputDevices(ctx)
	if err != nil {
		return fmt.Errorf("failed to list input devices: %w", err)
	}

	for _, device := range devices {
		if device.ID == deviceID && device.IsAvailable {
			return nil
		}
	}

	return ErrDeviceNotFound
}

// ValidateOutputDevice checks if an output device is available
func (d *DeviceDetector) ValidateOutputDevice(ctx context.Context, deviceID string) error {
	devices, err := d.ListOutputDevices(ctx)
	if err != nil {
		return fmt.Errorf("failed to list output devices: %w", err)
	}

	for _, device := range devices {
		if device.ID == deviceID && device.IsAvailable {
			return nil
		}
	}

	return ErrDeviceNotFound
}

// GetAudioSubsystem returns the appropriate audio subsystem for the platform
func (d *DeviceDetector) GetAudioSubsystem() string {
	switch d.platform {
	case "linux":
		// Check for PulseAudio first, then ALSA
		if d.isPulseAudioAvailable() {
			return "pulse"
		}
		return "alsa"
	case "darwin":
		return "avfoundation"
	case "windows":
		return "dshow"
	default:
		return "pulse" // Default fallback
	}
}

// isPulseAudioAvailable checks if PulseAudio is running
func (d *DeviceDetector) isPulseAudioAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "pulseaudio", "--check")
	return cmd.Run() == nil
}

// Linux-specific device detection
func (d *DeviceDetector) listLinuxInputDevices(ctx context.Context) ([]AudioDevice, error) {
	if d.isPulseAudioAvailable() {
		return d.listPulseAudioInputDevices(ctx)
	}
	return d.listALSAInputDevices(ctx)
}

func (d *DeviceDetector) listLinuxOutputDevices(ctx context.Context) ([]AudioDevice, error) {
	if d.isPulseAudioAvailable() {
		return d.listPulseAudioOutputDevices(ctx)
	}
	return d.listALSAOutputDevices(ctx)
}

func (d *DeviceDetector) listPulseAudioInputDevices(ctx context.Context) ([]AudioDevice, error) {
	cmd := exec.CommandContext(ctx, "pactl", "list", "short", "sources")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list PulseAudio sources: %w", err)
	}

	var devices []AudioDevice
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			devices = append(devices, AudioDevice{
				ID:          fields[1],
				Name:        fields[1],
				Type:        "input",
				IsDefault:   strings.Contains(fields[1], "default"),
				IsAvailable: true,
			})
		}
	}

	return devices, nil
}

func (d *DeviceDetector) listPulseAudioOutputDevices(ctx context.Context) ([]AudioDevice, error) {
	cmd := exec.CommandContext(ctx, "pactl", "list", "short", "sinks")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list PulseAudio sinks: %w", err)
	}

	var devices []AudioDevice
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			devices = append(devices, AudioDevice{
				ID:          fields[1],
				Name:        fields[1],
				Type:        "output",
				IsDefault:   strings.Contains(fields[1], "default"),
				IsAvailable: true,
			})
		}
	}

	return devices, nil
}

func (d *DeviceDetector) listALSAInputDevices(ctx context.Context) ([]AudioDevice, error) {
	// Basic ALSA device detection
	devices := []AudioDevice{
		{ID: "default", Name: "Default ALSA Input", Type: "input", IsDefault: true, IsAvailable: true},
		{ID: "hw:0,0", Name: "Hardware Device 0,0", Type: "input", IsDefault: false, IsAvailable: true},
	}
	return devices, nil
}

func (d *DeviceDetector) listALSAOutputDevices(ctx context.Context) ([]AudioDevice, error) {
	// Basic ALSA device detection
	devices := []AudioDevice{
		{ID: "default", Name: "Default ALSA Output", Type: "output", IsDefault: true, IsAvailable: true},
		{ID: "hw:0,0", Name: "Hardware Device 0,0", Type: "output", IsDefault: false, IsAvailable: true},
	}
	return devices, nil
}

// macOS-specific device detection
func (d *DeviceDetector) listMacOSInputDevices(ctx context.Context) ([]AudioDevice, error) {
	cmd := exec.CommandContext(ctx, "ffmpeg", "-f", "avfoundation", "-list_devices", "true", "-i", "")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// ffmpeg returns error when listing devices, but still provides output
	}

	devices := d.parseMacOSDevices(string(output), "input")
	return devices, nil
}

func (d *DeviceDetector) listMacOSOutputDevices(ctx context.Context) ([]AudioDevice, error) {
	cmd := exec.CommandContext(ctx, "ffmpeg", "-f", "avfoundation", "-list_devices", "true", "-i", "")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// ffmpeg returns error when listing devices, but still provides output
	}

	devices := d.parseMacOSDevices(string(output), "output")
	return devices, nil
}

func (d *DeviceDetector) parseMacOSDevices(output, deviceType string) []AudioDevice {
	var devices []AudioDevice
	lines := strings.Split(output, "\n")
	
	inSection := false
	sectionName := ""
	if deviceType == "input" {
		sectionName = "AVFoundation audio devices:"
	} else {
		sectionName = "AVFoundation audio devices:"
	}

	for _, line := range lines {
		if strings.Contains(line, sectionName) {
			inSection = true
			continue
		}
		
		if inSection && strings.Contains(line, "[") && strings.Contains(line, "]") {
			// Parse device line: [0] Built-in Microphone
			parts := strings.SplitN(line, "]", 2)
			if len(parts) == 2 {
				id := strings.Trim(parts[0], "[ ")
				name := strings.TrimSpace(parts[1])
				devices = append(devices, AudioDevice{
					ID:          id,
					Name:        name,
					Type:        deviceType,
					IsDefault:   strings.Contains(name, "Built-in"),
					IsAvailable: true,
				})
			}
		}
	}

	return devices
}

// Windows-specific device detection
func (d *DeviceDetector) listWindowsInputDevices(ctx context.Context) ([]AudioDevice, error) {
	cmd := exec.CommandContext(ctx, "ffmpeg", "-f", "dshow", "-list_devices", "true", "-i", "dummy")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// ffmpeg returns error when listing devices, but still provides output
	}

	devices := d.parseWindowsDevices(string(output), "input")
	return devices, nil
}

func (d *DeviceDetector) listWindowsOutputDevices(ctx context.Context) ([]AudioDevice, error) {
	cmd := exec.CommandContext(ctx, "ffmpeg", "-f", "dshow", "-list_devices", "true", "-i", "dummy")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// ffmpeg returns error when listing devices, but still provides output
	}

	devices := d.parseWindowsDevices(string(output), "output")
	return devices, nil
}

func (d *DeviceDetector) parseWindowsDevices(output, deviceType string) []AudioDevice {
	var devices []AudioDevice
	lines := strings.Split(output, "\n")
	
	inSection := false
	sectionName := ""
	if deviceType == "input" {
		sectionName = "DirectShow audio devices"
	} else {
		sectionName = "DirectShow audio devices"
	}

	for _, line := range lines {
		if strings.Contains(line, sectionName) {
			inSection = true
			continue
		}
		
		if inSection && strings.Contains(line, "\"") {
			// Parse device line: "Microphone (Realtek Audio)"
			start := strings.Index(line, "\"")
			end := strings.LastIndex(line, "\"")
			if start != -1 && end != -1 && start != end {
				deviceName := line[start+1 : end]
				devices = append(devices, AudioDevice{
					ID:          deviceName,
					Name:        deviceName,
					Type:        deviceType,
					IsDefault:   strings.Contains(deviceName, "Default"),
					IsAvailable: true,
				})
			}
		}
	}

	return devices
}