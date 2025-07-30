package minio_adapter

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// StorageConfig holds configuration for MinIO storage
type StorageConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	UseSSL          bool
	Region          string
}

// StorageOption is a functional option for configuring storage
type StorageOption func(*StorageConfig)

// WithEndpoint sets the MinIO endpoint
func WithEndpoint(endpoint string) StorageOption {
	return func(c *StorageConfig) {
		c.Endpoint = endpoint
	}
}

// WithCredentials sets the access credentials
func WithCredentials(accessKey, secretKey string) StorageOption {
	return func(c *StorageConfig) {
		c.AccessKeyID = accessKey
		c.SecretAccessKey = secretKey
	}
}

// WithBucket sets the bucket name
func WithBucket(bucket string) StorageOption {
	return func(c *StorageConfig) {
		c.BucketName = bucket
	}
}

// WithSSL enables or disables SSL
func WithSSL(useSSL bool) StorageOption {
	return func(c *StorageConfig) {
		c.UseSSL = useSSL
	}
}

// WithRegion sets the region
func WithRegion(region string) StorageOption {
	return func(c *StorageConfig) {
		c.Region = region
	}
}

// Storage implements the ObjectStore port using MinIO
type Storage struct {
	client *minio.Client
	config StorageConfig
	logger *slog.Logger
}

// NewStorage creates a new MinIO storage adapter
func NewStorage(logger *slog.Logger, opts ...StorageOption) (*Storage, error) {
	config := StorageConfig{
		Endpoint:        "localhost:9000",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		BucketName:      "speakr-audio",
		UseSSL:          false,
		Region:          "us-east-1",
	}

	for _, opt := range opts {
		opt(&config)
	}

	// Initialize MinIO client
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	storage := &Storage{
		client: client,
		config: config,
		logger: logger,
	}

	// Verify bucket exists
	if err := storage.ensureBucketExists(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	return storage, nil
}

// StoreAudio stores audio data in MinIO and returns the object path
func (s *Storage) StoreAudio(ctx context.Context, recordingID string, audioData io.Reader) (string, error) {
	logger := s.logger.With("recording_id", recordingID, "bucket", s.config.BucketName)

	objectName := fmt.Sprintf("recordings/%s.wav", recordingID)

	logger.Info("Storing audio file", "object_name", objectName)

	// Upload the audio file
	info, err := s.client.PutObject(ctx, s.config.BucketName, objectName, audioData, -1, minio.PutObjectOptions{
		ContentType: "audio/wav",
	})
	if err != nil {
		logger.Error("Failed to upload audio file", "error", err)
		
		// Check for specific error types
		if strings.Contains(err.Error(), "NoSuchBucket") {
			return "", ErrBucketNotFound
		}
		if strings.Contains(err.Error(), "AccessDenied") {
			return "", ErrAccessDenied
		}
		if strings.Contains(err.Error(), "InsufficientStorage") {
			return "", ErrInsufficientStorage
		}
		
		return "", fmt.Errorf("failed to store audio file: %w", err)
	}

	filePath := fmt.Sprintf("s3://%s/%s", s.config.BucketName, objectName)
	logger.Info("Audio file stored successfully", 
		"file_path", filePath, 
		"size", info.Size,
		"etag", info.ETag)

	return filePath, nil
}

// RetrieveAudio retrieves audio data from MinIO
func (s *Storage) RetrieveAudio(ctx context.Context, recordingID string) (io.Reader, error) {
	logger := s.logger.With("recording_id", recordingID, "bucket", s.config.BucketName)

	objectName := fmt.Sprintf("recordings/%s.wav", recordingID)

	logger.Info("Retrieving audio file", "object_name", objectName)

	// Get the object
	object, err := s.client.GetObject(ctx, s.config.BucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		logger.Error("Failed to retrieve audio file", "error", err)
		
		// Check for specific error types
		if strings.Contains(err.Error(), "NoSuchKey") {
			return nil, ErrObjectNotFound
		}
		if strings.Contains(err.Error(), "NoSuchBucket") {
			return nil, ErrBucketNotFound
		}
		if strings.Contains(err.Error(), "AccessDenied") {
			return nil, ErrAccessDenied
		}
		
		return nil, fmt.Errorf("failed to retrieve audio file: %w", err)
	}

	// Verify the object exists by getting its info
	_, err = object.Stat()
	if err != nil {
		logger.Error("Audio file not found", "error", err)
		return nil, ErrObjectNotFound
	}

	logger.Info("Audio file retrieved successfully")
	return object, nil
}

// ensureBucketExists checks if the bucket exists and creates it if necessary
func (s *Storage) ensureBucketExists(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.config.BucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		s.logger.Info("Bucket does not exist, creating it", "bucket", s.config.BucketName)
		
		err = s.client.MakeBucket(ctx, s.config.BucketName, minio.MakeBucketOptions{
			Region: s.config.Region,
		})
		if err != nil {
			return ErrBucketCreationFailed
		}
		
		s.logger.Info("Bucket created successfully", "bucket", s.config.BucketName)
	}

	return nil
}