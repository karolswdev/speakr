package openai_adapter

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// Embedder implements the EmbeddingGenerator port using OpenAI API
type Embedder struct {
	client *openai.Client
	model  string
}

// NewEmbedder creates a new OpenAI embedder
func NewEmbedder(apiKey, baseURL, model string) *Embedder {
	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}
	
	client := openai.NewClientWithConfig(config)
	
	return &Embedder{
		client: client,
		model:  model,
	}
}

// GenerateEmbedding generates a vector embedding for the given text
func (e *Embedder) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	req := openai.EmbeddingRequestStrings{
		Input: []string{text},
		Model: openai.EmbeddingModel(e.model),
	}

	resp, err := e.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	return resp.Data[0].Embedding, nil
}