package ports

import "context"

// SearchResult represents a single search result
type SearchResult struct {
	RecordingID      string                 `json:"recording_id"`
	TranscribedText  string                 `json:"transcribed_text"`
	Tags             []string               `json:"tags"`
	Metadata         map[string]interface{} `json:"metadata"`
	Similarity       float64                `json:"similarity"`
}

// SearchRequest represents a search query
type SearchRequest struct {
	QueryEmbedding []float32 `json:"query_embedding"`
	FilterTags     []string  `json:"filter_tags,omitempty"`
	Limit          int       `json:"limit,omitempty"`
}

// VectorSearcher defines the contract for searching vectors in the database
type VectorSearcher interface {
	Search(ctx context.Context, req SearchRequest) ([]SearchResult, error)
}