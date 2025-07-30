package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"speakr/embedder/internal/adapters/nats_adapter"
	"speakr/embedder/internal/adapters/openai_adapter"
	"speakr/embedder/internal/adapters/pgvector_adapter"
	"speakr/embedder/internal/core"

	"github.com/nats-io/nats.go"
)

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("Starting Embedding Service")

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

	// Create OpenAI embedder
	embedder, err := openai_adapter.NewEmbedder(logger,
		openai_adapter.WithAPIKey(config.OpenAIAPIKey),
		openai_adapter.WithBaseURL(config.OpenAIBaseURL),
		openai_adapter.WithModel(config.OpenAIModel),
		openai_adapter.WithTimeout(30*time.Second),
		openai_adapter.WithMaxRetries(3),
	)
	if err != nil {
		logger.Error("Failed to create OpenAI embedder", "error", err)
		os.Exit(1)
	}

	// Create PostgreSQL vector store
	vectorStore, err := pgvector_adapter.NewStore(logger,
		pgvector_adapter.WithHost(config.DBHost),
		pgvector_adapter.WithPort(config.DBPort),
		pgvector_adapter.WithCredentials(config.DBUser, config.DBPassword),
		pgvector_adapter.WithDatabase(config.DBName),
		pgvector_adapter.WithSSLMode("disable"),
		pgvector_adapter.WithMaxConnections(10),
		pgvector_adapter.WithTimeout(30*time.Second),
	)
	if err != nil {
		logger.Error("Failed to create vector store", "error", err)
		os.Exit(1)
	}
	defer vectorStore.Close()

	// Create core service
	service := core.NewService(embedder, vectorStore, logger)

	// Create NATS subscriber
	subscriber := nats_adapter.NewSubscriber(natsConn, logger)
	defer subscriber.Close()

	// Subscribe to transcription.succeeded events
	err = subscriber.Subscribe(ctx, "speakr.event.transcription.succeeded", service.HandleTranscriptionEvent)
	if err != nil {
		logger.Error("Failed to subscribe to transcription events", "error", err)
		os.Exit(1)
	}

	// Start health check server
	go startHealthServer(logger, config.HealthPort)

	logger.Info("Embedding Service started successfully")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down Embedding Service")
	cancel()

	// Give some time for graceful shutdown
	time.Sleep(2 * time.Second)
	logger.Info("Embedding Service stopped")
}

// Config holds the service configuration
type Config struct {
	NatsURL       string
	OpenAIAPIKey  string
	OpenAIBaseURL string
	OpenAIModel   string
	DBHost        string
	DBPort        int
	DBUser        string
	DBPassword    string
	DBName        string
	HealthPort    string
}

func loadConfig() (*Config, error) {
	config := &Config{
		NatsURL:       getEnvOrDefault("NATS_URL", "nats://localhost:4222"),
		OpenAIAPIKey:  os.Getenv("OPENAI_API_KEY"),
		OpenAIBaseURL: getEnvOrDefault("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		OpenAIModel:   getEnvOrDefault("OPENAI_EMBEDDING_MODEL", "text-embedding-ada-002"),
		DBHost:        getEnvOrDefault("DB_HOST", "localhost"),
		DBUser:        getEnvOrDefault("DB_USER", "postgres"),
		DBPassword:    getEnvOrDefault("DB_PASSWORD", "postgres"),
		DBName:        getEnvOrDefault("DB_NAME", "speakr"),
		HealthPort:    getEnvOrDefault("HEALTH_PORT", "8081"),
	}

	// Parse DB port
	dbPortStr := getEnvOrDefault("DB_PORT", "5432")
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}
	config.DBPort = dbPort

	// Validate required fields
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

func validateBaseURL(baseURL string) error {
	if baseURL == "" {
		return fmt.Errorf("base URL cannot be empty")
	}
	
	// Basic URL validation
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		return fmt.Errorf("base URL must start with http:// or https://")
	}
	
	return nil
}

func startHealthServer(logger *slog.Logger, port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"embedder"}`))
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