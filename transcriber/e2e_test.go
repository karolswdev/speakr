package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"speakr/transcriber/internal/core"

	"github.com/nats-io/nats.go"
)

// End-to-end test to verify complete transcriber service functionality
func TestE2E_CompleteTranscriptionFlow(t *testing.T) {
	// Skip if not in integration test mode
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping E2E test (set INTEGRATION_TEST=true to run)")
	}

	// Connect to NATS for sending commands and receiving events
	natsConn, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		t.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer natsConn.Close()

	// Subscribe to events to verify the complete flow
	eventChan := make(chan *nats.Msg, 10)
	
	// Subscribe to all event subjects
	eventSubjects := []string{
		"speakr.event.recording.started",
		"speakr.event.recording.finished",
		"speakr.event.recording.cancelled",
		"speakr.event.transcription.succeeded",
		"speakr.event.transcription.failed",
	}

	var subs []*nats.Subscription
	for _, subject := range eventSubjects {
		sub, err := natsConn.Subscribe(subject, func(msg *nats.Msg) {
			eventChan <- msg
		})
		if err != nil {
			t.Fatalf("Failed to subscribe to %s: %v", subject, err)
		}
		subs = append(subs, sub)
	}

	// Cleanup subscriptions
	defer func() {
		for _, sub := range subs {
			sub.Unsubscribe()
		}
	}()

	t.Log("Testing complete recording and transcription flow...")

	// Test 1: Start Recording Command
	startCmd := core.StartRecordingCommand{
		OutputFormat: "wav",
		Tags:         []string{"e2e-test", "recording"},
		Metadata:     map[string]interface{}{"test_type": "e2e", "timestamp": time.Now().Unix()},
	}

	startData, err := json.Marshal(startCmd)
	if err != nil {
		t.Fatalf("Failed to marshal start command: %v", err)
	}

	err = natsConn.Publish("speakr.command.recording.start", startData)
	if err != nil {
		t.Fatalf("Failed to publish start command: %v", err)
	}

	t.Log("Published recording start command")

	// Wait for recording.started event
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var recordingID string
	select {
	case msg := <-eventChan:
		if msg.Subject == "speakr.event.recording.started" {
			var event map[string]interface{}
			if err := json.Unmarshal(msg.Data, &event); err != nil {
				t.Fatalf("Failed to unmarshal recording.started event: %v", err)
			}
			recordingID = event["recording_id"].(string)
			t.Logf("Received recording.started event for recording ID: %s", recordingID)
		} else {
			t.Fatalf("Expected recording.started event, got %s", msg.Subject)
		}
	case <-ctx.Done():
		t.Fatal("Timeout waiting for recording.started event")
	}

	// Test 2: Stop Recording Command (with transcription)
	stopCmd := core.StopRecordingCommand{
		RecordingID:      recordingID,
		TranscribeOnStop: true,
		Metadata:         map[string]interface{}{"test_type": "e2e"},
	}

	stopData, err := json.Marshal(stopCmd)
	if err != nil {
		t.Fatalf("Failed to marshal stop command: %v", err)
	}

	err = natsConn.Publish("speakr.command.recording.stop", stopData)
	if err != nil {
		t.Fatalf("Failed to publish stop command: %v", err)
	}

	t.Log("Published recording stop command")

	// Wait for recording.finished event
	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()

	select {
	case msg := <-eventChan:
		if msg.Subject == "speakr.event.recording.finished" {
			var event map[string]interface{}
			if err := json.Unmarshal(msg.Data, &event); err != nil {
				t.Fatalf("Failed to unmarshal recording.finished event: %v", err)
			}
			t.Logf("Received recording.finished event with file path: %s", event["audio_file_path"])
		} else {
			t.Fatalf("Expected recording.finished event, got %s", msg.Subject)
		}
	case <-ctx2.Done():
		t.Fatal("Timeout waiting for recording.finished event")
	}

	// Since transcribe_on_stop was true, we should also get a transcription event
	// (It will likely fail since we're not actually recording real audio, but we can verify the flow)
	ctx3, cancel3 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel3()

	select {
	case msg := <-eventChan:
		if msg.Subject == "speakr.event.transcription.succeeded" {
			var event map[string]interface{}
			if err := json.Unmarshal(msg.Data, &event); err != nil {
				t.Fatalf("Failed to unmarshal transcription.succeeded event: %v", err)
			}
			t.Logf("Received transcription.succeeded event with text: %s", event["transcribed_text"])
		} else if msg.Subject == "speakr.event.transcription.failed" {
			var event map[string]interface{}
			if err := json.Unmarshal(msg.Data, &event); err != nil {
				t.Fatalf("Failed to unmarshal transcription.failed event: %v", err)
			}
			t.Logf("Received transcription.failed event (expected with mock audio): %s", event["error"])
		} else {
			t.Fatalf("Expected transcription event, got %s", msg.Subject)
		}
	case <-ctx3.Done():
		t.Fatal("Timeout waiting for transcription event")
	}

	t.Log("E2E test completed successfully - all events received in correct order")
}

func TestE2E_DirectTranscriptionCommand(t *testing.T) {
	// Skip if not in integration test mode
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping E2E test (set INTEGRATION_TEST=true to run)")
	}

	// Connect to NATS
	natsConn, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		t.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer natsConn.Close()

	// Subscribe to transcription events
	eventChan := make(chan *nats.Msg, 5)
	sub, err := natsConn.Subscribe("speakr.event.transcription.*", func(msg *nats.Msg) {
		eventChan <- msg
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to transcription events: %v", err)
	}
	defer sub.Unsubscribe()

	t.Log("Testing direct transcription command with base64 audio data...")

	// Test direct transcription with base64 audio data
	transcribeCmd := core.TranscriptionCommand{
		AudioData: "UklGRiQAAABXQVZFZm10IBAAAAABAAEARKwAAIhYAQACABAAZGF0YQAAAAA=", // Minimal WAV header in base64
		Tags:      []string{"e2e-test", "direct-transcription"},
		Metadata:  map[string]interface{}{"test_type": "direct"},
	}

	transcribeData, err := json.Marshal(transcribeCmd)
	if err != nil {
		t.Fatalf("Failed to marshal transcription command: %v", err)
	}

	err = natsConn.Publish("speakr.command.transcription.run", transcribeData)
	if err != nil {
		t.Fatalf("Failed to publish transcription command: %v", err)
	}

	t.Log("Published direct transcription command")

	// Wait for transcription result
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	select {
	case msg := <-eventChan:
		if strings.HasSuffix(msg.Subject, ".succeeded") {
			var event map[string]interface{}
			if err := json.Unmarshal(msg.Data, &event); err != nil {
				t.Fatalf("Failed to unmarshal transcription event: %v", err)
			}
			t.Logf("Received transcription.succeeded event: %s", event["transcribed_text"])
		} else if strings.HasSuffix(msg.Subject, ".failed") {
			var event map[string]interface{}
			if err := json.Unmarshal(msg.Data, &event); err != nil {
				t.Fatalf("Failed to unmarshal transcription event: %v", err)
			}
			t.Logf("Received transcription.failed event (expected with minimal audio): %s", event["error"])
		}
	case <-ctx.Done():
		t.Fatal("Timeout waiting for transcription event")
	}

	t.Log("Direct transcription test completed successfully")
}