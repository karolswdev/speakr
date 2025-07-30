package openai_adapter

import "errors"

var (
	// ErrAPIKeyNotSet indicates the OpenAI API key is not configured
	ErrAPIKeyNotSet = errors.New("OpenAI API key not set")
	
	// ErrInvalidModel indicates an invalid embedding model was specified
	ErrInvalidModel = errors.New("invalid embedding model")
	
	// ErrAPICallFailed indicates the OpenAI API call failed
	ErrAPICallFailed = errors.New("OpenAI API call failed")
)