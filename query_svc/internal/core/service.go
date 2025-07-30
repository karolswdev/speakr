package core

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/speakr/query_svc/internal/ports"
)

// QueryRequest represents an incoming query request
type QueryRequest struct {
	QueryText  string   `json:"query_text"`
	FilterTags []string `json:"filter_tags,omitempty"`
	Limit      int      `json:"limit,omitempty"`
}

// Service implements the core query service logic
type Service struct {
	embeddingGenerator ports.EmbeddingGenerator
	vectorSearcher     ports.VectorSearcher
	logger             *slog.Logger
}

// NewService creates a new query service instance
func NewService(embeddingGenerator ports.EmbeddingGenerator, vectorSearcher ports.VectorSearcher, logger *slog.Logger) *Service {
	return &Service{
		embeddingGenerator: embeddingGenerator,
		vectorSearcher:     vectorSearcher,
		logger:             logger,
	}
}

// Search performs a semantic search on the vector database
func (s *Service) Search(ctx context.Context, req QueryRequest) ([]ports.SearchResult, error) {
	correlationID := ctx.Value("correlation_id")
	
	s.logger.Info("Starting search operation",
		"correlation_id", correlationID,
		"query_text_length", len(req.QueryText),
		"filter_tags", req.FilterTags,
		"limit", req.Limit,
	)

	// Validate input
	if strings.TrimSpace(req.QueryText) == "" {
		s.logger.Error("Invalid query text provided",
			"correlation_id", correlationID,
		)
		return nil, ErrInvalidQuery
	}

	// Set default limit if not provided
	if req.Limit <= 0 {
		req.Limit = 10
	}

	// Generate embedding for the query text
	embedding, err := s.embeddingGenerator.GenerateEmbedding(ctx, req.QueryText)
	if err != nil {
		s.logger.Error("Failed to generate embedding",
			"correlation_id", correlationID,
			"error", err,
		)
		return nil, fmt.Errorf("%w: %v", ErrEmbeddingFailed, err)
	}

	s.logger.Info("Generated embedding for query",
		"correlation_id", correlationID,
		"embedding_dimensions", len(embedding),
	)

	// Perform vector search
	searchReq := ports.SearchRequest{
		QueryEmbedding: embedding,
		FilterTags:     req.FilterTags,
		Limit:          req.Limit,
	}

	results, err := s.vectorSearcher.Search(ctx, searchReq)
	if err != nil {
		s.logger.Error("Failed to perform vector search",
			"correlation_id", correlationID,
			"error", err,
		)
		return nil, fmt.Errorf("%w: %v", ErrSearchFailed, err)
	}

	s.logger.Info("Search completed successfully",
		"correlation_id", correlationID,
		"results_count", len(results),
	)

	return results, nil
}