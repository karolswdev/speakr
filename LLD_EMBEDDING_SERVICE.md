# LLD: The Embedding Service

**ID:** LLD-ES
**Version:** 1.0
**Status:** Proposed

---

### 1. Overview

This document provides the low-level design for the **Embedding Service**.

-   **Purpose:** To listen for successful transcription events, generate vector embeddings for the transcribed text, and store the result in a vector database for semantic search.
-   **Architectural Adherence:** This service **must** strictly adhere to all principles outlined in `ARCHITECTURE.md`.

### 2. Core Components & Packages

The service will be structured into the following Go packages within the `embedder/` directory:

-   `cmd/`: The main entry point for the service.
-   `internal/core`: The core application logic.
-   `internal/adapters`:
    -   `nats_adapter/`: Implements the NATS subscriber.
    -   `openai_adapter/`: Implements the `EmbeddingGenerator` port by calling the OpenAI Embeddings API.
    -   `pgvector_adapter/`: Implements the `VectorStore` port for writing data to the PostgreSQL/pgvector database.
-   `internal/ports`: Defines the Go interfaces for `EmbeddingGenerator` and `VectorStore`.

### 3. Logic Flow

#### 3.1. On `speakr.event.transcription.succeeded`

1.  The `nats_adapter` receives the event from the NATS bus.
2.  It calls the `core` service's `ProcessTranscription` method, passing the `transcribed_text`, `recording_id`, and `tags`.
3.  The `core` service invokes the `openai_adapter` to generate a vector embedding from the `transcribed_text`.
4.  The `core` service then commands the `pgvector_adapter` to save the complete record:
    -   `recording_id` (as primary key)
    -   `transcribed_text`
    -   `tags`
    -   The generated `embedding` vector
5.  (Optional) The service may publish a `speakr.event.embedding.succeeded` event for further downstream processing or logging.

### 4. Configuration (Environment Variables)

-   `NATS_URL`: URL for the NATS server.
-   `OPENAI_API_KEY`: API key for the OpenAI service.
-   `DB_HOST`: Hostname for the PostgreSQL database.
-   `DB_PORT`: Port for the PostgreSQL database.
-   `DB_USER`: Username for the PostgreSQL database.
-   `DB_PASSWORD`: Password for the PostgreSQL database.
-   `DB_NAME`: Name of the database to use.
