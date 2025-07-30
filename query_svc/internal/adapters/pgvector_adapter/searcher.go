package pgvector_adapter

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/speakr/query_svc/internal/ports"
)

// Searcher implements the VectorSearcher port using PostgreSQL with pgvector
type Searcher struct {
	db *sql.DB
}

// NewSearcher creates a new PostgreSQL vector searcher
func NewSearcher(db *sql.DB) *Searcher {
	return &Searcher{
		db: db,
	}
}

// Search performs a vector similarity search in the database
func (s *Searcher) Search(ctx context.Context, req ports.SearchRequest) ([]ports.SearchResult, error) {
	// Build the SQL query with optional tag filtering
	query := `
		SELECT 
			recording_id,
			transcribed_text,
			tags,
			metadata,
			1 - (embedding <=> $1) as similarity
		FROM embeddings
	`
	
	args := []interface{}{pq.Array(req.QueryEmbedding)}
	argIndex := 2

	// Add tag filtering if provided
	if len(req.FilterTags) > 0 {
		placeholders := make([]string, len(req.FilterTags))
		for i, tag := range req.FilterTags {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, tag)
			argIndex++
		}
		query += fmt.Sprintf(" WHERE tags && ARRAY[%s]", strings.Join(placeholders, ","))
	}

	// Order by similarity and limit results
	query += fmt.Sprintf(" ORDER BY similarity DESC LIMIT $%d", argIndex)
	args = append(args, req.Limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var results []ports.SearchResult
	for rows.Next() {
		var result ports.SearchResult
		var tagsArray pq.StringArray
		var metadataJSON []byte

		err := rows.Scan(
			&result.RecordingID,
			&result.TranscribedText,
			&tagsArray,
			&metadataJSON,
			&result.Similarity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}

		// Convert tags array
		result.Tags = []string(tagsArray)

		// Parse metadata JSON
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &result.Metadata); err != nil {
				return nil, fmt.Errorf("failed to parse metadata JSON: %w", err)
			}
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating search results: %w", err)
	}

	return results, nil
}