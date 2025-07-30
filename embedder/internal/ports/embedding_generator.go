package ports

import (
	"context"
)

// EmbeddingGenerator defines the interface for generating vector embeddings from text
type EmbeddingGenerator interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
}