package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/speakr/query_svc/internal/adapters/http_adapter"
	"github.com/speakr/query_svc/internal/adapters/openai_adapter"
	"github.com/speakr/query_svc/internal/adapters/pgvector_adapter"
	"github.com/speakr/query_svc/internal/core"
)

// Config holds the application configuration
type Config struct {
	HTTPPort         string
	OpenAIAPIKey     string
	OpenAIBaseURL    string
	OpenAIModel      string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
}

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("Starting Query Service")

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Setup database connection
	db, err := setupDatabase(config)
	if err != nil {
		logger.Error("Failed to setup database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create adapters
	embeddingGenerator := openai_adapter.NewEmbedder(
		config.OpenAIAPIKey,
		config.OpenAIBaseURL,
		config.OpenAIModel,
	)

	vectorSearcher := pgvector_adapter.NewSearcher(db)

	// Create core service
	service := core.NewService(embeddingGenerator, vectorSearcher, logger)

	// Create HTTP handler
	handler := http_adapter.NewHandler(service, logger)

	// Setup HTTP server
	server := &http.Server{
		Addr:    ":" + config.HTTPPort,
		Handler: handler.SetupRoutes(),
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", "port", config.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("Server exited")
}

// loadConfig loads configuration from environment variables
func loadConfig() (*Config, error) {
	config := &Config{
		HTTPPort:      getEnvOrDefault("HTTP_PORT", "8080"),
		OpenAIAPIKey:  os.Getenv("OPENAI_API_KEY"),
		OpenAIBaseURL: os.Getenv("OPENAI_BASE_URL"),
		OpenAIModel:   getEnvOrDefault("OPENAI_EMBEDDING_MODEL", "text-embedding-ada-002"),
		DBHost:        getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:        getEnvOrDefault("DB_PORT", "5432"),
		DBUser:        getEnvOrDefault("DB_USER", "postgres"),
		DBPassword:    os.Getenv("DB_PASSWORD"),
		DBName:        getEnvOrDefault("DB_NAME", "speakr"),
	}

	// Validate required configuration
	if config.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	if config.DBPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD environment variable is required")
	}

	return config, nil
}

// setupDatabase creates and configures the database connection
func setupDatabase(config *Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// getEnvOrDefault returns the environment variable value or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}