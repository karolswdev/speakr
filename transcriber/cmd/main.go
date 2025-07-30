package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"speakr/transcriber/internal/adapters/ffmpeg_adapter"
	"speakr/transcriber/internal/adapters/minio_adapter"
	"speakr/transcriber/internal/adapters/nats_adapter"
	"speakr/transcriber/internal/adapters/openai_adapter"
	"speakr/transcriber/internal/core"

	"github.com/nats-io/nats.go"
)

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("Starting Transcriber Service")

	// Load configuration from environment
	config, err := loadConfig()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to NATS
	natsConn, err := nats.Connect(config.NatsURL)
	if err != nil {
		logger.Error("Failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	defer natsConn.Close()

	logger.Info("Connected to NATS", "url", config.NatsURL)

	// Create real adapters
	audioRecorder, err := ffmpeg_adapter.NewRecorder(logger,
		ffmpeg_adapter.WithTempDir("/tmp/speakr"),
		ffmpeg_adapter.WithSampleRate(44100),
		ffmpeg_adapter.WithChannels(1),
	)
	if err != nil {
		logger.Error("Failed to create audio recorder", "error", err)
		os.Exit(1)
	}

	objectStore, err := minio_adapter.NewStorage(logger,
		minio_adapter.WithEndpoint(config.MinioEndpoint),
		minio_adapter.WithCredentials(config.MinioAccessKey, config.MinioSecretKey),
		minio_adapter.WithBucket(config.MinioBucketName),
		minio_adapter.WithSSL(false),
	)
	if err != nil {
		logger.Error("Failed to create object store", "error", err)
		os.Exit(1)
	}

	// Create real OpenAI transcription service
	transcriptionSvc, err := openai_adapter.NewTranscriber(logger,
		openai_adapter.WithAPIKey(config.OpenAIAPIKey),
		openai_adapter.WithBaseURL(config.OpenAIBaseURL),
		openai_adapter.WithModel(config.OpenAITranscriptionModel),
		openai_adapter.WithTimeout(30*time.Second),
		openai_adapter.WithMaxRetries(3),
	)
	if err != nil {
		logger.Error("Failed to create transcription service", "error", err)
		os.Exit(1)
	}

	eventPublisher := nats_adapter.NewPublisher(natsConn, logger)

	// Create core service
	service := core.NewService(
		audioRecorder,
		transcriptionSvc,
		objectStore,
		eventPublisher,
		logger,
	)

	// Create and start NATS subscriber
	subscriber := nats_adapter.NewSubscriber(natsConn, service, logger)
	if err := subscriber.Subscribe(ctx); err != nil {
		logger.Error("Failed to setup NATS subscriptions", "error", err)
		os.Exit(1)
	}

	// Start health check server
	go startHealthServer(logger, config.HealthPort)

	logger.Info("Transcriber Service started successfully")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down Transcriber Service")
	cancel()

	// Give some time for graceful shutdown
	time.Sleep(2 * time.Second)
	logger.Info("Transcriber Service stopped")
}

// Config holds the service configuration
type Config struct {
	NatsURL                 string
	OpenAIAPIKey            string
	OpenAIBaseURL           string
	OpenAITranscriptionModel string
	MinioEndpoint           string
	MinioAccessKey          string
	MinioSecretKey          string
	MinioBucketName         string
	HealthPort              string
}

func loadConfig() (*Config, error) {
	config := &Config{
		NatsURL:                 getEnvOrDefault("NATS_URL", "nats://localhost:4222"),
		OpenAIAPIKey:            os.Getenv("OPENAI_API_KEY"),
		OpenAIBaseURL:           getEnvOrDefault("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		OpenAITranscriptionModel: getEnvOrDefault("OPENAI_TRANSCRIPTION_MODEL", "whisper-1"),
		MinioEndpoint:           getEnvOrDefault("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey:          getEnvOrDefault("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey:          getEnvOrDefault("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucketName:         getEnvOrDefault("MINIO_BUCKET_NAME", "speakr-audio"),
		HealthPort:              getEnvOrDefault("HEALTH_PORT", "8080"),
	}

	if config.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	// Validate base URL format
	if err := validateBaseURL(config.OpenAIBaseURL); err != nil {
		return nil, fmt.Errorf("invalid OPENAI_BASE_URL: %w", err)
	}

	return config, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func startHealthServer(logger *slog.Logger, port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"transcriber"}`))
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	logger.Info("Health server starting", "port", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Health server failed", "error", err)
	}
}
func validateBaseURL(baseURL string) error {
	if baseURL == "" {
		return fmt.Errorf("base URL cannot be empty")
	}
	
	// Basic URL validation
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		return fmt.Errorf("base URL must start with http:// or https://")
	}
	
	// Remove trailing slash for consistency
	if strings.HasSuffix(baseURL, "/") {
		baseURL = strings.TrimSuffix(baseURL, "/")
	}
	
	return nil
}
