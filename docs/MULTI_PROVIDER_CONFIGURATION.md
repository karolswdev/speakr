# Multi-Provider Configuration Guide

The Speakr platform supports using different OpenAI-compatible providers for transcription and embedding services. This enables cost optimization, performance tuning, and vendor diversity.

## Configuration Overview

### Shared Configuration (Default)
```bash
# Used by both services if service-specific variables are not set
OPENAI_API_KEY=your-shared-api-key
OPENAI_BASE_URL=https://api.openai.com/v1
```

### Service-Specific Configuration

#### Transcriber Service
```bash
# Uses shared OPENAI_API_KEY and OPENAI_BASE_URL if not specified
OPENAI_TRANSCRIPTION_MODEL=whisper-1
```

#### Embedding Service
```bash
# Optional: Override shared configuration for embeddings
OPENAI_EMBEDDING_API_KEY=your-embedding-specific-key
OPENAI_EMBEDDING_BASE_URL=https://api.openai.com/v1
OPENAI_EMBEDDING_MODEL=text-embedding-ada-002
```

## Multi-Provider Examples

### Example 1: OpenAI for Transcription, Groq for Embeddings
```bash
# Transcription via OpenAI
OPENAI_API_KEY=sk-your-openai-key
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_TRANSCRIPTION_MODEL=whisper-1

# Embeddings via Groq (faster, cheaper)
OPENAI_EMBEDDING_API_KEY=gsk_your-groq-key
OPENAI_EMBEDDING_BASE_URL=https://api.groq.com/openai/v1
OPENAI_EMBEDDING_MODEL=text-embedding-ada-002
```

### Example 2: Azure OpenAI for Transcription, Local Ollama for Embeddings
```bash
# Transcription via Azure OpenAI
OPENAI_API_KEY=your-azure-key
OPENAI_BASE_URL=https://your-resource.openai.azure.com/openai/deployments/whisper/v1
OPENAI_TRANSCRIPTION_MODEL=whisper-1

# Embeddings via local Ollama
OPENAI_EMBEDDING_API_KEY=ollama
OPENAI_EMBEDDING_BASE_URL=http://localhost:11434/v1
OPENAI_EMBEDDING_MODEL=nomic-embed-text
```

### Example 3: Different Models, Same Provider
```bash
# Shared OpenAI account with different models
OPENAI_API_KEY=sk-your-openai-key
OPENAI_BASE_URL=https://api.openai.com/v1

# High-accuracy transcription
OPENAI_TRANSCRIPTION_MODEL=whisper-large

# Cost-effective embeddings
OPENAI_EMBEDDING_MODEL=text-embedding-3-small
```

### Example 4: Completely Separate Providers
```bash
# Transcription via Groq (fast)
OPENAI_API_KEY=gsk_your-groq-key
OPENAI_BASE_URL=https://api.groq.com/openai/v1
OPENAI_TRANSCRIPTION_MODEL=whisper-large-v3

# Embeddings via Together.ai (cost-effective)
OPENAI_EMBEDDING_API_KEY=your-together-key
OPENAI_EMBEDDING_BASE_URL=https://api.together.xyz/v1
OPENAI_EMBEDDING_MODEL=togethercomputer/m2-bert-80M-8k-retrieval
```

## Configuration Fallback Logic

1. **Embedding Service**: 
   - Uses `OPENAI_EMBEDDING_*` variables if set
   - Falls back to shared `OPENAI_*` variables
   - Ensures backward compatibility

2. **Transcriber Service**:
   - Uses shared `OPENAI_API_KEY` and `OPENAI_BASE_URL`
   - Uses `OPENAI_TRANSCRIPTION_MODEL` for model selection

## Provider-Specific Considerations

### Model Compatibility
- **OpenAI**: `whisper-1`, `text-embedding-ada-002`, `text-embedding-3-small`, `text-embedding-3-large`
- **Groq**: `whisper-large-v3`, `text-embedding-ada-002`
- **Azure OpenAI**: Deployment-specific model names
- **Ollama**: Local model names like `nomic-embed-text`, `whisper`
- **Together.ai**: Provider-specific model identifiers

### Rate Limits and Costs
- **OpenAI**: Standard rate limits, premium pricing
- **Groq**: Very fast inference, competitive pricing
- **Azure OpenAI**: Enterprise controls, predictable pricing
- **Ollama**: No rate limits (local), no API costs
- **Together.ai**: High throughput, cost-effective

## Testing Multi-Provider Setup

### Test Transcription Configuration
```bash
# Test with specific transcription model
OPENAI_TRANSCRIPTION_MODEL=whisper-large ./bin/transcriber
```

### Test Embedding Configuration
```bash
# Test with separate embedding provider
OPENAI_EMBEDDING_API_KEY=different-key \
OPENAI_EMBEDDING_BASE_URL=https://api.groq.com/openai/v1 \
./bin/embedder
```

### Integration Testing
```bash
# Test complete pipeline with mixed providers
export OPENAI_API_KEY=sk-transcription-key
export OPENAI_EMBEDDING_API_KEY=gsk-embedding-key
export OPENAI_EMBEDDING_BASE_URL=https://api.groq.com/openai/v1
export INTEGRATION_TEST=true

# Run end-to-end test
make test-integration
```

## Troubleshooting

### Common Issues
1. **Model Not Found**: Verify the model name is supported by the provider
2. **Authentication Errors**: Check API key format matches provider requirements
3. **Rate Limiting**: Different providers have different rate limits
4. **Response Format**: Ensure provider returns OpenAI-compatible responses

### Debugging
```bash
# Enable debug logging
export LOG_LEVEL=debug

# Check configuration
./bin/transcriber --config-check
./bin/embedder --config-check
```

## Best Practices

1. **Cost Optimization**: Use faster/cheaper providers for embeddings, higher quality for transcription
2. **Redundancy**: Configure fallback providers for critical workloads
3. **Testing**: Always test provider combinations before production deployment
4. **Monitoring**: Track costs and performance across different providers
5. **Security**: Use separate API keys for different services when possible