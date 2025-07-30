package pgvector_adapter

import (
	"testing"
)

func TestNewSearcher(t *testing.T) {
	// This is a unit test for the constructor
	// Integration tests with real database would be run separately
	
	searcher := NewSearcher(nil) // nil DB for unit test
	
	if searcher == nil {
		t.Fatal("Expected searcher to be created, got nil")
	}
}

// Note: Integration tests with real PostgreSQL database would be run separately
// with INTEGRATION_TEST=true environment variable