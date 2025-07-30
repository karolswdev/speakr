package pgvector_adapter

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"speakr/embedder/internal/ports"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// StoreConfig holds configuration for the PostgreSQL store
type StoreConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxConns int
	Timeout  time.Duration
}

// StoreOption is a functional option for configuring the store
type StoreOption func(*StoreConfig)

// WithHost sets the database host
func WithHost(host string) StoreOption {
	return func(c *StoreConfig) {
		c.Host = host
	}
}

// WithPort sets the database port
func WithPort(port int) StoreOption {
	return func(c *StoreConfig) {
		c.Port = port
	}
}

// WithCredentials sets the database credentials
func WithCredentials(user, password string) StoreOption {
	return func(c *StoreConfig) {
		c.User = user
		c.Password = password
	}
}

// WithDatabase sets the database name
func WithDatabase(dbName string) StoreOption {
	return func(c *StoreConfig) {
		c.DBName = dbName
	}
}

// WithSSLMode sets the SSL mode
func WithSSLMode(sslMode string) StoreOption {
	return func(c *StoreConfig) {
		c.SSLMode = sslMode
	}
}

// WithMaxConnections sets the maximum number of connections
func WithMaxConnections(maxConns int) StoreOption {
	return func(c *StoreConfig) {
		c.MaxConns = maxConns
	}
}

// WithTimeout sets the connection timeout
func WithTimeout(timeout time.Duration) StoreOption {
	return func(c *StoreConfig) {
		c.Timeout = timeout
	}
}

// Store implements the VectorStore port using PostgreSQL with pgvector
type Store struct {
	db     *sql.DB
	config StoreConfig
	logger *slog.Logger
}

// NewStore creates a new PostgreSQL vector store
func NewStore(logger *slog.Logger, opts ...StoreOption) (*Store, error) {
	config := StoreConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "speakr",
		SSLMode:  "disable",
		MaxConns: 10,
		Timeout:  30 * time.Second,
	}

	for _, opt := range opts {
		opt(&config)
	}

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxConns)
	db.SetMaxIdleConns(config.MaxConns / 2)
	db.SetConnMaxLifetime(time.Hour)

	store := &Store{
		db:     db,
		config: config,
		logger: logger,
	}

	// Test connection and ensure table exists
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	if err := store.ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := store.ensureTableExists(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ensure table exists: %w", err)
	}

	return store, nil
}

// StoreRecord stores a vector record in the database
func (s *Store) StoreRecord(ctx context.Context, record ports.VectorRecord) error {
	logger := s.logger.With(
		"recording_id", record.RecordingID,
		"operation", "store_record",
	)

	logger.Info("Storing vector record", 
		"text_length", len(record.TranscribedText),
		"embedding_dimensions", len(record.Embedding),
		"tags_count", len(record.Tags))

	// Validate input
	if record.RecordingID == "" {
		return ErrMissingRecordingID
	}
	if record.TranscribedText == "" {
		return ErrEmptyText
	}
	if len(record.Embedding) == 0 {
		return ErrEmptyEmbedding
	}

	// Convert embedding to pgvector format
	embeddingStr := vectorToString(record.Embedding)

	// Insert or update record
	query := `
		INSERT INTO transcriptions (recording_id, transcribed_text, tags, embedding, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (recording_id) 
		UPDATE SET 
			transcribed_text = EXCLUDED.transcribed_text,
			tags = EXCLUDED.tags,
			embedding = EXCLUDED.embedding,
			updated_at = NOW()
	`

	_, err := s.db.ExecContext(ctx, query, 
		record.RecordingID, 
		record.TranscribedText, 
		pq.Array(record.Tags), 
		embeddingStr)
	
	if err != nil {
		logger.Error("Failed to store vector record", "error", err)
		
		// Check for specific error types
		if strings.Contains(err.Error(), "connection") {
			return ErrDatabaseUnavailable
		}
		if strings.Contains(err.Error(), "constraint") {
			return ErrInvalidData
		}
		
		return fmt.Errorf("failed to store vector record: %w", err)
	}

	logger.Info("Vector record stored successfully")
	return nil
}

// GetRecord retrieves a vector record from the database
func (s *Store) GetRecord(ctx context.Context, recordingID string) (*ports.VectorRecord, error) {
	logger := s.logger.With(
		"recording_id", recordingID,
		"operation", "get_record",
	)

	logger.Info("Retrieving vector record")

	if recordingID == "" {
		return nil, ErrMissingRecordingID
	}

	query := `
		SELECT recording_id, transcribed_text, tags, embedding
		FROM transcriptions 
		WHERE recording_id = $1
	`

	var record ports.VectorRecord
	var embeddingStr string
	var tags pq.StringArray

	err := s.db.QueryRowContext(ctx, query, recordingID).Scan(
		&record.RecordingID,
		&record.TranscribedText,
		&tags,
		&embeddingStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warn("Vector record not found")
			return nil, nil // Not found, but not an error
		}
		
		logger.Error("Failed to retrieve vector record", "error", err)
		
		if strings.Contains(err.Error(), "connection") {
			return nil, ErrDatabaseUnavailable
		}
		
		return nil, fmt.Errorf("failed to retrieve vector record: %w", err)
	}

	// Convert tags and embedding
	record.Tags = []string(tags)
	record.Embedding = stringToVector(embeddingStr)

	logger.Info("Vector record retrieved successfully",
		"text_length", len(record.TranscribedText),
		"embedding_dimensions", len(record.Embedding),
		"tags_count", len(record.Tags))

	return &record, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// ping tests the database connection
func (s *Store) ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// ensureTableExists creates the transcriptions table if it doesn't exist
func (s *Store) ensureTableExists(ctx context.Context) error {
	s.logger.Info("Ensuring transcriptions table exists")

	// Enable pgvector extension
	_, err := s.db.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		return fmt.Errorf("failed to create vector extension: %w", err)
	}

	// Create table with proper indexes
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS transcriptions (
			recording_id UUID PRIMARY KEY,
			transcribed_text TEXT NOT NULL,
			tags TEXT[] DEFAULT '{}',
			embedding vector(1536) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`

	_, err = s.db.ExecContext(ctx, createTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create transcriptions table: %w", err)
	}

	// Create indexes for performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_transcriptions_tags ON transcriptions USING GIN (tags)",
		"CREATE INDEX IF NOT EXISTS idx_transcriptions_text ON transcriptions USING GIN (to_tsvector('english', transcribed_text))",
		"CREATE INDEX IF NOT EXISTS idx_transcriptions_embedding ON transcriptions USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100)",
		"CREATE INDEX IF NOT EXISTS idx_transcriptions_created_at ON transcriptions (created_at)",
	}

	for _, indexQuery := range indexes {
		_, err = s.db.ExecContext(ctx, indexQuery)
		if err != nil {
			s.logger.Warn("Failed to create index", "error", err, "query", indexQuery)
			// Continue with other indexes even if one fails
		}
	}

	s.logger.Info("Transcriptions table and indexes ensured")
	return nil
}

// vectorToString converts a float32 slice to pgvector string format
func vectorToString(vector []float32) string {
	if len(vector) == 0 {
		return "[]"
	}
	
	parts := make([]string, len(vector))
	for i, v := range vector {
		parts[i] = fmt.Sprintf("%g", v)
	}
	
	return "[" + strings.Join(parts, ",") + "]"
}

// stringToVector converts a pgvector string to float32 slice
func stringToVector(s string) []float32 {
	// Remove brackets and split by comma
	s = strings.Trim(s, "[]")
	if s == "" {
		return []float32{}
	}
	
	parts := strings.Split(s, ",")
	vector := make([]float32, len(parts))
	
	for i, part := range parts {
		var f float32
		fmt.Sscanf(strings.TrimSpace(part), "%g", &f)
		vector[i] = f
	}
	
	return vector
}