package minio_adapter

import "errors"

// Custom error types for predictable failures
var (
	ErrBucketNotFound        = errors.New("specified bucket does not exist")
	ErrBucketCreationFailed  = errors.New("failed to create bucket")
	ErrObjectNotFound        = errors.New("specified object does not exist")
	ErrAccessDenied          = errors.New("access denied to MinIO resource")
	ErrInsufficientStorage   = errors.New("insufficient storage space available")
	ErrConnectionFailed      = errors.New("failed to connect to MinIO server")
	ErrInvalidCredentials    = errors.New("invalid MinIO credentials")
)