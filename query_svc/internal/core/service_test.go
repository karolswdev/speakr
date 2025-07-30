package core

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/speakr/query_svc/internal/ports"
)

// Mock implementations for testing
type mockEmbeddingGenerator struct {
	embedding []float32
	err       error
}

func (m *mockEmbeddingGenerator) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.embedding, nil
}

type mockVectorSearcher struct {
	results []ports.SearchResult
	err     error
}

func (m *mockVectorSearcher) Search(ctx context.Context, req ports.SearchRequest) ([]ports.SearchResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.results, nil
}

func TestService_Search_Success(t *testing.T) {
	// Setup mocks
	mockEmbedding := []float32{0.1, 0.2, 0.3}
	mockResults := []ports.SearchResult{
		{
			RecordingID:     "rec-123",
			TranscribedText: "test transcription",
			Tags:            []string{"tag1"},
			Metadata:        map[string]interface{}{"key": "value"},
			Similarity:      0.95,
		},
	}

	embeddingGen := &mockEmbeddingGenerator{embedding: mockEmbedding}
	vectorSearcher := &mockVectorSearcher{results: mockResults}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := NewService(embeddingGen, vectorSearcher, logger)

	// Test successful search
	ctx := context.WithValue(context.Background(), "correlation_id", "test-123")
	req := QueryRequest{
		QueryText:  "test query",
		FilterTags: []string{"tag1"},
		Limit:      10,
	}

	results, err := service.Search(ctx, req)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got: %d", len(results))
	}

	if results[0].RecordingID != "rec-123" {
		t.Errorf("Expected recording_id 'rec-123', got: %s", results[0].RecordingID)
	}
}

func TestService_Search_InvalidQuery(t *testing.T) {
	embeddingGen := &mockEmbeddingGenerator{}
	vectorSearcher := &mockVectorSearcher{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := NewService(embeddingGen, vectorSearcher, logger)

	ctx := context.Background()
	req := QueryRequest{
		QueryText: "   ", // Empty/whitespace query
		Limit:     10,
	}

	_, err := service.Search(ctx, req)

	if err != ErrInvalidQuery {
		t.Errorf("Expected ErrInvalidQuery, got: %v", err)
	}
}

func TestService_Search_EmbeddingFailure(t *testing.T) {
	embeddingGen := &mockEmbeddingGenerator{err: errors.New("API failure")}
	vectorSearcher := &mockVectorSearcher{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := NewService(embeddingGen, vectorSearcher, logger)

	ctx := context.Background()
	req := QueryRequest{
		QueryText: "test query",
		Limit:     10,
	}

	_, err := service.Search(ctx, req)

	if !errors.Is(err, ErrEmbeddingFailed) {
		t.Errorf("Expected ErrEmbeddingFailed, got: %v", err)
	}
}

func TestService_Search_SearchFailure(t *testing.T) {
	embeddingGen := &mockEmbeddingGenerator{embedding: []float32{0.1, 0.2}}
	vectorSearcher := &mockVectorSearcher{err: errors.New("DB failure")}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := NewService(embeddingGen, vectorSearcher, logger)

	ctx := context.Background()
	req := QueryRequest{
		QueryText: "test query",
		Limit:     10,
	}

	_, err := service.Search(ctx, req)

	if !errors.Is(err, ErrSearchFailed) {
		t.Errorf("Expected ErrSearchFailed, got: %v", err)
	}
}

func TestService_Search_DefaultLimit(t *testing.T) {
	mockEmbedding := []float32{0.1, 0.2, 0.3}
	embeddingGen := &mockEmbeddingGenerator{embedding: mockEmbedding}
	
	// Create a custom mock that captures the request
	capturedReq := &ports.SearchRequest{}
	vectorSearcher := &mockVectorSearcherWithCapture{
		results:     []ports.SearchResult{},
		capturedReq: capturedReq,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(embeddingGen, vectorSearcher, logger)

	ctx := context.Background()
	req := QueryRequest{
		QueryText: "test query",
		// Limit not set, should default to 10
	}

	_, err := service.Search(ctx, req)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if capturedReq.Limit != 10 {
		t.Errorf("Expected default limit 10, got: %d", capturedReq.Limit)
	}
}

type mockVectorSearcherWithCapture struct {
	results     []ports.SearchResult
	err         error
	capturedReq *ports.SearchRequest
}

func (m *mockVectorSearcherWithCapture) Search(ctx context.Context, req ports.SearchRequest) ([]ports.SearchResult, error) {
	*m.capturedReq = req
	if m.err != nil {
		return nil, m.err
	}
	return m.results, nil
}