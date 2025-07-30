package nats_adapter

import (
	"encoding/json"
	"testing"

	"speakr/transcriber/internal/core"
)

func TestHandleStartRecording_ValidJSON(t *testing.T) {
	// Create test command
	cmd := core.StartRecordingCommand{
		OutputFormat: "wav",
		Tags:         []string{"test"},
		Metadata:     map[string]interface{}{"source": "test"},
	}
	
	data, err := json.Marshal(cmd)
	if err != nil {
		t.Fatalf("Failed to marshal command: %v", err)
	}
	
	// Test JSON unmarshaling
	var parsedCmd core.StartRecordingCommand
	err = json.Unmarshal(data, &parsedCmd)
	if err != nil {
		t.Fatalf("Failed to unmarshal command: %v", err)
	}
	
	// Verify the command was parsed correctly
	if parsedCmd.OutputFormat != "wav" {
		t.Errorf("Expected output format 'wav', got %s", parsedCmd.OutputFormat)
	}
	
	if len(parsedCmd.Tags) != 1 || parsedCmd.Tags[0] != "test" {
		t.Errorf("Expected tags ['test'], got %v", parsedCmd.Tags)
	}
}

func TestHandleStartRecording_InvalidJSON(t *testing.T) {
	// Test with invalid JSON
	invalidJSON := []byte(`{"invalid": json}`)
	
	var cmd core.StartRecordingCommand
	err := json.Unmarshal(invalidJSON, &cmd)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestContractCompliance_AllCommands(t *testing.T) {
	// Test all command types can be marshaled/unmarshaled correctly
	
	// Start recording command
	startCmd := core.StartRecordingCommand{
		OutputFormat: "wav",
		Tags:         []string{"project-x", "daily-standup"},
		Metadata:     map[string]interface{}{"triggered_by": "cli-adapter"},
	}
	
	data, err := json.Marshal(startCmd)
	if err != nil {
		t.Fatalf("Failed to marshal start command: %v", err)
	}
	
	var parsedStartCmd core.StartRecordingCommand
	if err := json.Unmarshal(data, &parsedStartCmd); err != nil {
		t.Fatalf("Failed to unmarshal start command: %v", err)
	}
	
	// Stop recording command
	stopCmd := core.StopRecordingCommand{
		RecordingID:      "a1b2c3d4-e5f6-...",
		TranscribeOnStop: true,
		Metadata:         map[string]interface{}{"copy_to_clipboard": true},
	}
	
	data, err = json.Marshal(stopCmd)
	if err != nil {
		t.Fatalf("Failed to marshal stop command: %v", err)
	}
	
	var parsedStopCmd core.StopRecordingCommand
	if err := json.Unmarshal(data, &parsedStopCmd); err != nil {
		t.Fatalf("Failed to unmarshal stop command: %v", err)
	}
	
	// Cancel recording command
	cancelCmd := core.CancelRecordingCommand{
		RecordingID: "a1b2c3d4-e5f6-...",
	}
	
	data, err = json.Marshal(cancelCmd)
	if err != nil {
		t.Fatalf("Failed to marshal cancel command: %v", err)
	}
	
	var parsedCancelCmd core.CancelRecordingCommand
	if err := json.Unmarshal(data, &parsedCancelCmd); err != nil {
		t.Fatalf("Failed to unmarshal cancel command: %v", err)
	}
	
	// Transcription command
	transcribeCmd := core.TranscriptionCommand{
		RecordingID: "a1b2c3d4-e5f6-...",
		Tags:        []string{"additional-tag"},
		Metadata:    map[string]interface{}{"copy_to_clipboard": true},
	}
	
	data, err = json.Marshal(transcribeCmd)
	if err != nil {
		t.Fatalf("Failed to marshal transcribe command: %v", err)
	}
	
	var parsedTranscribeCmd core.TranscriptionCommand
	if err := json.Unmarshal(data, &parsedTranscribeCmd); err != nil {
		t.Fatalf("Failed to unmarshal transcribe command: %v", err)
	}
}