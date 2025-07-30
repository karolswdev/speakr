# LLD: The Query Service

**ID:** LLD-QS
**Version:** 1.0
**Status:** Proposed

---

### 1. Overview

This document provides the low-level design for the **Query Service**.

-   **Purpose:** To provide a public-facing API that allows users or RAG frameworks to perform semantic searches on the transcribed and vectorized data.
-   **Architectural Adherence:** This service **must** strictly adhere to all principles outlined in `ARCHITECTURE.md`.

### 2. Core Components & Packages

The service will be structured into the following Go packages within the `query_svc/` directory:

-   `cmd/`: The main entry point for the service.
-   `internal/core`: The core application logic.
-   `internal/adapters`:
    -   `http_adapter/`: Implements the public-facing REST API (e.g., using `net/http` or a lightweight framework like `chi`).
    -   `openai_adapter/`: Implements the `EmbeddingGenerator` port to vectorize the incoming query text.
    -   `pgvector_adapter/`: Implements the `VectorSearcher` port to query the vector database.
-   `internal/ports`: Defines the Go interfaces for `EmbeddingGenerator` and `VectorSearcher`.

### 3. Logic Flow

#### 3.1. On `POST /api/v1/query`

1.  The `http_adapter` receives an incoming API request. The request body contains `query_text` and optional `filter_tags`.
2.  The adapter calls the `core` service's `Search` method.
3.  The `core` service first sends the `query_text` to the `openai_adapter` to get its vector embedding.
4.  The `core` service then passes the generated embedding and the `filter_tags` to the `pgvector_adapter`.
5.  The `pgvector_adapter` executes a query against the database to find the top N most similar vectors, filtering by the provided tags.
6.  The `pgvector_adapter` returns the search results (including the original text and metadata) to the `core` service.
7.  The `core` service returns the results to the `http_adapter`, which formats them as a JSON response and sends it back to the client.

### 4. Configuration (Environment Variables)

-   `HTTP_PORT`: The port on which to run the HTTP server (e.g., `8080`).
-   `OPENAI_API_KEY`: API key for the OpenAI service.
-   `DB_HOST`: Hostname for the PostgreSQL database.
-   `DB_PORT`: Port for the PostgreSQL database.
-   `DB_USER`: Username for the PostgreSQL database.
-   `DB_PASSWORD`: Password for the PostgreSQL database.
-   `DB_NAME`: Name of the database to use.
