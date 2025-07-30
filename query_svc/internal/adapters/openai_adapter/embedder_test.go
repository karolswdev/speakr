package openai_adapter

import (
	"testing"
)

func TestNewEmbedder(t *testing.T) {
	embedder := NewEmbedder("test-key", "https://api.openai.com/v1", "text-embedding-ada-002")
	
	if embedder == nil {
		t.Fatal("Expected embedder to be created, got nil")
	}
	
	if embedder.model != "text-embedding-ada-002" {
		t.Errorf("Expected model 'text-embedding-ada-002', got: %s", embedder.model)
	}
}

func TestNewEmbedder_WithCustomBaseURL(t *testing.T) {
	customURL := "https://custom-api.example.com/v1"
	embedder := NewEmbedder("test-key", customURL, "custom-model")
	
	if embedder == nil {
		t.Fatal("Expected embedder to be created, got nil")
	}
	
	if embedder.model != "custom-model" {
		t.Errorf("Expected model 'custom-model', got: %s", embedder.model)
	}
}

// Note: Integration tests with real OpenAI API would be run separately
// with INTEGRATION_TEST=true environment variable