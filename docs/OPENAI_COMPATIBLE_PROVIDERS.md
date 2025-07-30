# OpenAI-Compatible Provider Configuration

The Speakr Transcriber Service supports any OpenAI-compatible API provider through the `OPENAI_BASE_URL` environment variable. This allows you to use alternative providers for cost optimization, performance, or vendor diversity.

## Supported Providers

### OpenAI (Default)
```bash
OPENAI_API_KEY=sk-your-openai-key
OPENAI_BASE_URL=https://api.openai.com/v1
```

### Groq (High-speed inference)
```bash
OPENAI_API_KEY=gsk_your-groq-key
OPENAI_BASE_URL=https://api.groq.com/openai/v1
```

### Azure OpenAI
```bash
OPENAI_API_KEY=your-azure-key
OPENAI_BASE_URL=https://your-resource.openai.azure.com/openai/deployments/your-deployment/v1
```

### Ollama (Local deployment)
```bash
OPENAI_API_KEY=ollama  # Can be any value for local Ollama
OPENAI_BASE_URL=http://localhost:11434/v1
```

### Together.ai
```bash
OPENAI_API_KEY=your-together-key
OPENAI_BASE_URL=https://api.together.xyz/v1
```

### Anyscale Endpoints
```bash
OPENAI_API_KEY=your-anyscale-key
OPENAI_BASE_URL=https://api.endpoints.anyscale.com/v1
```

## Configuration Notes

1. **API Key**: Each provider requires their own API key format. Some local providers like Ollama may accept any value.

2. **Model Compatibility**: The service uses the `whisper-1` model by default. Ensure your chosen provider supports this model or a compatible speech-to-text model.

3. **Rate Limits**: Different providers have different rate limits. Adjust the `WithMaxRetries` and `WithTimeout` settings in the service configuration as needed.

4. **Response Format**: All providers should return responses in the standard OpenAI format:
   ```json
   {
     "text": "transcribed text here"
   }
   ```

## Testing Provider Compatibility

To test a new provider, you can run the integration test:

```bash
export OPENAI_API_KEY=your-provider-key
export OPENAI_BASE_URL=https://your-provider-endpoint/v1
export INTEGRATION_TEST=true
cd transcriber
go test -v -run TestTranscribeAudio_Integration
```

## Provider-Specific Considerations

### Groq
- Extremely fast inference
- Limited model selection
- May have different rate limiting behavior

### Azure OpenAI
- Requires specific deployment URL format
- May have different authentication requirements
- Enterprise-grade security and compliance

### Ollama
- Runs locally, no external API calls
- Requires local model installation
- No API key validation needed

### Together.ai
- Multiple model options
- Competitive pricing
- Good for high-volume usage

## Troubleshooting

1. **Invalid Base URL**: Ensure the URL starts with `http://` or `https://`
2. **Authentication Errors**: Verify the API key format matches the provider's requirements
3. **Model Not Found**: Check if the provider supports the `whisper-1` model
4. **Timeout Issues**: Increase the timeout value for slower providers
5. **Rate Limiting**: Implement exponential backoff or reduce concurrent requests

## Adding New Providers

To add support for a new OpenAI-compatible provider:

1. Test the provider's API compatibility with the OpenAI format
2. Update this documentation with configuration examples
3. Add any provider-specific error handling if needed
4. Test with the integration test suite