-- Speakr Database Initialization Script
-- This script sets up the database schema for the Embedding and Query services
-- per LLD-ES and LLD-QS specifications

-- Enable the pgvector extension for vector similarity search
CREATE EXTENSION IF NOT EXISTS vector;

-- Create the main table for storing transcriptions and their embeddings
-- This supports the Embedding Service (LLD-ES) and Query Service (LLD-QS)
CREATE TABLE IF NOT EXISTS transcriptions (
    -- Primary key using recording_id from the CONTRACT.md
    recording_id UUID PRIMARY KEY,
    
    -- The transcribed text from the Transcriber Service
    transcribed_text TEXT NOT NULL,
    
    -- Tags array for filtering and organization per CONTRACT.md
    tags TEXT[] DEFAULT '{}',
    
    -- Vector embedding for semantic search (1536 dimensions for OpenAI text-embedding-ada-002)
    embedding VECTOR(1536),
    
    -- Timestamps for auditing
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for efficient querying
-- Index for vector similarity search (cosine distance)
CREATE INDEX IF NOT EXISTS idx_transcriptions_embedding_cosine 
ON transcriptions USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

-- Index for tag-based filtering
CREATE INDEX IF NOT EXISTS idx_transcriptions_tags 
ON transcriptions USING GIN (tags);

-- Index for text search
CREATE INDEX IF NOT EXISTS idx_transcriptions_text 
ON transcriptions USING GIN (to_tsvector('english', transcribed_text));

-- Index for timestamp-based queries
CREATE INDEX IF NOT EXISTS idx_transcriptions_created_at 
ON transcriptions (created_at DESC);

-- Create a function to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_transcriptions_updated_at 
    BEFORE UPDATE ON transcriptions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Insert a test record to verify the setup
INSERT INTO transcriptions (recording_id, transcribed_text, tags, embedding) 
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'This is a test transcription to verify the database setup.',
    ARRAY['test', 'setup'],
    -- Generate a random vector for testing (in production, this comes from OpenAI)
    (SELECT ARRAY(SELECT random() FROM generate_series(1, 1536)))::vector
) ON CONFLICT (recording_id) DO NOTHING;

-- Create a view for easy querying without exposing the vector directly
CREATE OR REPLACE VIEW transcription_summary AS
SELECT 
    recording_id,
    transcribed_text,
    tags,
    created_at,
    updated_at,
    -- Calculate vector magnitude for debugging
    vector_dims(embedding) as embedding_dimensions
FROM transcriptions;

-- Grant necessary permissions (adjust as needed for production)
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO postgres;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO postgres;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO postgres;

-- Display setup completion message
DO $$
BEGIN
    RAISE NOTICE 'Speakr database initialization completed successfully!';
    RAISE NOTICE 'Created tables: transcriptions';
    RAISE NOTICE 'Created indexes: embedding (ivfflat), tags (GIN), text (GIN), timestamps';
    RAISE NOTICE 'Enabled extensions: vector';
    RAISE NOTICE 'Test record inserted with recording_id: 00000000-0000-0000-0000-000000000001';
END $$;