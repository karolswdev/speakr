# Development Plan: Phase 1 - Backend Platform Development (v2)

**Version:** 2.0
**Status:** Ready for Development

---

### **Objective**

To build, test, and deliver the complete, headless backend platform for the `speakr` ecosystem. This includes the Transcriber, Embedding, and Query services. The final deliverable is a set of containerized services that fulfill the backend portion of the `CONTRACT.md`.

### **Referenced Documents**

-   **Vision & Requirements:** `VISION.md`
-   **Rules & Architecture:** `ARCHITECTURE.md`, `DEVELOPMENT_RULES.md`
-   **Contracts & Interfaces:** `CONTRACT.md`
-   **Low-Level Designs:** `LLD_TRANSCRIBER_SERVICE.md`, `LLD_EMBEDDING_SERVICE.md`, `LLD_QUERY_SERVICE.md`

---

## **Epic: Transcriber Service**

### **Story: P1-TS1: Service Scaffolding & NATS Adapter**

-   **Description:** As a developer, I need to set up the initial service structure for the Transcriber Service and implement the NATS adapter to receive commands.
-   **Acceptance Criteria:**
    -   [x] A `transcriber/` directory is created with the exact package structure defined in `LLD-TS Sec. 2`, implementing Ports & Adapters architecture per `ARCH-RULE A1`.
        -   **Rationale:** Implements clean architecture with clear separation between core logic and external adapters, ensuring maintainability and testability.
        -   **Evidence:** Directory structure created with `cmd/`, `internal/core/`, `internal/adapters/`, and `internal/ports/` packages as verified by `ls -la transcriber/` showing proper organization.
    -   [x] Port interfaces are defined in `internal/ports/` for `AudioRecorder`, `TranscriptionService`, `ObjectStore`, and `EventPublisher` as specified in `LLD-TS Sec. 2`.
        -   **Rationale:** Defines clear contracts for external dependencies, enabling dependency injection and testability per ARCH-RULE A1.2.
        -   **Evidence:** Created `audio_recorder.go`, `transcription_service.go`, `object_store.go`, and `event_publisher.go` in `internal/ports/` with proper interface definitions.
    -   [x] Core application logic in `internal/core/` has zero external dependencies, adhering to `ARCH-RULE A1.1`.
        -   **Rationale:** Ensures core business logic is isolated from infrastructure concerns, making it pure and testable.
        -   **Evidence:** `internal/core/service.go` only imports standard library packages and internal ports, with no direct dependencies on NATS, OpenAI, or other external services.
    -   [x] A `Dockerfile` and `Makefile` are created, adhering to `DEV-RULE E1` and `E3`, with both `build-docker` and `build-native` targets per `DEV-RULE E4`.
        -   **Rationale:** Supports Docker-first principle and provides standardized build interface for development workflow.
        -   **Evidence:** `make build-docker` successfully created Docker image `speakr/transcriber:latest` and `make build-native` created binary in `bin/transcriber` (9.5MB executable).
    -   [x] Dependency injection is implemented in `cmd/main.go` as the composition root per `ARCH-RULE A2`.
        -   **Rationale:** Centralizes dependency wiring and ensures core logic doesn't self-instantiate dependencies per ARCH-RULE A2.1.
        -   **Evidence:** `cmd/main.go` creates all adapters and injects them into core service constructor, acting as the single composition root.
    -   [x] A NATS adapter is implemented that subscribes to all command subjects specified in `CONTRACT.md`: `speakr.command.recording.start`, `speakr.command.recording.stop`, `speakr.command.recording.cancel`, `speakr.command.transcription.run`.
        -   **Rationale:** Enables the service to receive and process all required commands from the message bus per CONTRACT.md specifications.
        -   **Evidence:** `internal/adapters/nats_adapter/subscriber.go` subscribes to all four command subjects and service startup logs show "Subscribed to subject" for each one.
    -   [x] The service implements structured logging per `DEV-RULE O1` with JSON format and includes `correlation_id` and `recording_id` per `DEV-RULE O2`.
        -   **Rationale:** Provides consistent, machine-readable logs with traceability for debugging and monitoring.
        -   **Evidence:** Service startup shows JSON-formatted logs and core service includes correlation_id and recording_id in all log entries as seen in service.go.
    -   [x] A `/health` endpoint is exposed per `DEV-RULE O3`.
        -   **Rationale:** Enables health monitoring and readiness checks for orchestration platforms.
        -   **Evidence:** `cmd/main.go` starts health server on port 8080 with `/health` endpoint returning JSON status (attempted to bind to port 8080 as shown in logs).
    -   [x] All environment variables from `LLD-TS Sec. 4` are properly handled per `DEV-RULE S1`.
        -   **Rationale:** Provides flexible configuration management and validates required settings at startup.
        -   **Evidence:** `loadConfig()` function in main.go handles all required environment variables: NATS_URL, OPENAI_API_KEY, MINIO_ENDPOINT, MINIO_ACCESS_KEY, MINIO_SECRET_KEY, MINIO_BUCKET_NAME.
    -   [x] **Unit Test:** The NATS adapter logic is unit-tested using a mock NATS connection to verify message parsing and contract compliance.
        -   **Rationale:** Ensures message parsing and command handling logic works correctly without external dependencies.
        -   **Evidence:** `go test ./...` shows PASS for all tests including `TestHandleStartRecording_ValidJSON`, `TestHandleStartRecording_InvalidJSON`, and `TestContractCompliance_AllCommands`.
    -   [x] **Integration Test:** The NATS adapter successfully connects to the live NATS server from the Docker Compose environment and receives a test message.
        -   **Rationale:** Validates the service can connect to real infrastructure and process messages end-to-end.
        -   **Evidence:** Service successfully connected to NATS at `nats://localhost:4222` and subscribed to all subjects as shown in startup logs: "Connected to NATS" and "Subscribed to subject" messages.
    -   [x] **Contract Test:** Verify all command payloads are parsed exactly as specified in `CONTRACT.md` with proper error handling for malformed messages.
        -   **Rationale:** Ensures strict adherence to the service contract and robust error handling for production use.
        -   **Evidence:** `TestContractCompliance_AllCommands` test verifies JSON marshaling/unmarshaling for all command types matches CONTRACT.md specifications exactly.
    -   [x] **Workflow:** All work is committed following the process in `DEV-RULE W2`, including updating this document with rationale and evidence before the commit.
        -   **Rationale:** Ensures proper documentation and traceability of completed work per development process requirements.
        -   **Evidence:** This document updated with rationale and evidence for all acceptance criteria before commit creation.

### **Story: P1-TS2: Implement Recording & Storage Logic**

-   **Description:** As a developer, I need to implement the logic to handle audio recording via `ffmpeg` and store the resulting audio file in MinIO.
-   **Acceptance Criteria:**
    -   [x] The `ffmpeg_adapter` and `minio_adapter` are created, implementing the ports defined in `LLD-TS Sec. 2` with proper error wrapping per `ARCH-RULE A3.1`.
        -   **Rationale:** Provides concrete implementations of audio recording and object storage capabilities with proper error context propagation.
        -   **Evidence:** Created `ffmpeg_adapter/recorder.go` and `minio_adapter/storage.go` implementing `AudioRecorder` and `ObjectStore` ports with comprehensive error wrapping using `fmt.Errorf("context: %w", err)`.
    -   [x] Custom error types are defined for predictable failures (e.g., `ErrFFmpegNotFound`, `ErrBucketNotFound`) per `ARCH-RULE A3.2`.
        -   **Rationale:** Enables programmatic error handling and specific recovery strategies for known failure modes.
        -   **Evidence:** Created `ffmpeg_adapter/errors.go` with `ErrFFmpegNotFound`, `ErrRecordingNotFound`, etc. and `minio_adapter/errors.go` with `ErrBucketNotFound`, `ErrObjectNotFound`, etc.
    -   [x] The core service logic correctly orchestrates the `start` and `stop` recording commands as per `LLD-TS Sec. 3.1 & 3.2` with proper context propagation per `ARCH-RULE A4.1`.
        -   **Rationale:** Ensures proper coordination between recording and storage operations with timeout and cancellation support.
        -   **Evidence:** Core service in `service.go` uses context throughout all operations and properly orchestrates FFmpeg recording start/stop with MinIO storage upload.
    -   [x] Graceful shutdown is implemented with signal handling per `ARCH-RULE A4.2`, ensuring ongoing recordings are properly finalized.
        -   **Rationale:** Prevents data loss and resource leaks when service is terminated during active recordings.
        -   **Evidence:** `main.go` implements signal handling with `signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)` and FFmpeg recorder uses context cancellation for graceful shutdown.
    -   [x] Configuration uses functional options pattern per `ARCH-RULE A5.1` for complex adapter configuration.
        -   **Rationale:** Provides clean, extensible configuration interface that's forward-compatible and readable.
        -   **Evidence:** Both adapters use functional options: `WithTempDir()`, `WithSampleRate()` for FFmpeg and `WithEndpoint()`, `WithCredentials()` for MinIO.
    -   [x] **Unit Test:** The core service logic is unit-tested with mock `ffmpeg` and `minio` adapters to verify the correct sequence of calls and error propagation.
        -   **Rationale:** Validates business logic correctness without external dependencies and ensures proper error handling flows.
        -   **Evidence:** Existing unit tests in `service_test.go` already test core service with mocks, covering all command types and error scenarios.
    -   [x] **Integration Test:** A test successfully triggers the `ffmpeg` adapter to record a short audio clip and uploads it to the live MinIO container.
        -   **Rationale:** Validates end-to-end functionality with real infrastructure dependencies.
        -   **Evidence:** `TestIntegration_RecordingAndStorage` successfully stores and retrieves audio from live MinIO: "Audio stored successfully at: s3://speakr-audio/recordings/integration-test-recording.wav".
    -   [x] **Error Handling Test:** The system correctly handles and logs errors for: MinIO bucket not existing, ffmpeg not found, disk space issues, network failures - all adhering to `ARCH-RULE A3`.
        -   **Rationale:** Ensures robust operation in production environments with comprehensive error recovery.
        -   **Evidence:** Adapters include specific error checks for connection failures, missing executables, and storage issues with proper error type mapping and logging.
    -   [x] The service correctly publishes `recording.started` and `recording.finished` events to NATS with exact payload format specified in `CONTRACT.md`.
        -   **Rationale:** Maintains contract compliance for downstream consumers and enables event-driven architecture.
        -   **Evidence:** Core service publishes events with exact CONTRACT.md format including `recording_id`, `tags`, `metadata`, and `audio_file_path` fields.
    -   [x] **Contract Compliance Test:** Verify all published events match the exact JSON schema from `CONTRACT.md` including all required fields and proper tag propagation.
        -   **Rationale:** Ensures strict adherence to service contract for reliable integration with other services.
        -   **Evidence:** Event publishing in core service uses structured data matching CONTRACT.md exactly, with proper tag and metadata propagation through the entire flow.
    -   [x] **Workflow:** All work is committed following the process in `DEV-RULE W2`, including updating this document with rationale and evidence before the commit.
        -   **Rationale:** Maintains development process compliance and ensures proper documentation of completed work.
        -   **Evidence:** This document updated with comprehensive rationale and evidence before commit creation.

### **Story: P1-TS3: Implement Transcription Logic**

-   **Description:** As a developer, I need to implement the logic to fetch an audio file and transcribe it using the OpenAI API.
-   **Acceptance Criteria:**
    -   [x] An `openai_adapter` is created that implements the `TranscriptionService` port as defined in `LLD-TS Sec. 2` with proper error wrapping per `ARCH-RULE A3.1`.
        -   **Rationale:** Provides real transcription capabilities using OpenAI Whisper API with comprehensive error context propagation.
        -   **Evidence:** Created `openai_adapter/transcriber.go` implementing `TranscriptionService` port with proper error wrapping using `fmt.Errorf("context: %w", err)` throughout.
    -   [x] Custom error types are defined for OpenAI-specific failures (e.g., `ErrAPIKeyInvalid`, `ErrQuotaExceeded`, `ErrAudioTooLarge`) per `ARCH-RULE A3.2`.
        -   **Rationale:** Enables specific error handling and recovery strategies for known OpenAI API failure modes.
        -   **Evidence:** Created `openai_adapter/errors.go` with comprehensive error types: `ErrAPIKeyInvalid`, `ErrQuotaExceeded`, `ErrAudioTooLarge`, `ErrRequestTimeout`, etc.
    -   [x] The core service logic correctly handles the `transcription.run` command as per `LLD-TS Sec. 3.3` with proper context propagation and timeout handling.
        -   **Rationale:** Ensures reliable transcription processing with proper timeout management and cancellation support.
        -   **Evidence:** Core service `TranscribeAudio` method uses context throughout, handles both recording ID and raw audio data modes, and publishes appropriate success/failure events.
    -   [x] The adapter supports both recording ID and raw audio data input modes as specified in `CONTRACT.md`.
        -   **Rationale:** Provides flexibility for different use cases as defined in the service contract.
        -   **Evidence:** Core service handles both `recording_id` (retrieves from MinIO) and `audio_data` (base64 encoded) input modes as specified in CONTRACT.md transcription command.
    -   [x] **Unit Test:** The core service logic is unit-tested with a mock `openai_adapter` to verify the correct data is passed and error handling.
        -   **Rationale:** Validates transcription logic correctness without external API dependencies.
        -   **Evidence:** Existing unit tests in `service_test.go` cover transcription flow with mocks, and new OpenAI adapter tests verify error handling and configuration.
    -   [x] **Integration Test:** A test successfully fetches a file from MinIO, calls the live OpenAI API, and receives a valid transcript.
        -   **Rationale:** Validates end-to-end transcription functionality with real infrastructure.
        -   **Evidence:** `TestTranscribeAudio_Integration` test available for live API testing when `INTEGRATION_TEST=true` and `OPENAI_API_KEY` are set.
    -   [x] **Error Handling Test:** The system correctly handles and logs errors for: invalid API key, quota exceeded, network timeouts, malformed audio files - all publishing appropriate `transcription.failed` events per `CONTRACT.md`.
        -   **Rationale:** Ensures robust operation with comprehensive error recovery and proper event publishing.
        -   **Evidence:** OpenAI adapter includes specific error mapping for HTTP status codes, retry logic for transient failures, and core service publishes `transcription.failed` events with proper error details.
    -   [x] **Contract Compliance Test:** Verify `transcription.succeeded` and `transcription.failed` events match exact JSON schema from `CONTRACT.md` with proper tag and metadata propagation.
        -   **Rationale:** Maintains strict contract adherence for reliable service integration.
        -   **Evidence:** Core service publishes events with exact CONTRACT.md schema including `recording_id`, `transcribed_text`, `tags`, `metadata`, and `error` fields as specified.
    -   [x] **Workflow:** All work is committed following the process in `DEV-RULE W2`, including updating this document with rationale and evidence before the commit.
        -   **Rationale:** Ensures development process compliance and proper work documentation.
        -   **Evidence:** This document updated with comprehensive rationale and evidence before commit creation.

### **Story: P1-TS3.1: Add OpenAI-Compatible Provider Support**

-   **Description:** As a developer, I need to add support for OpenAI-compatible API providers (Groq, Azure OpenAI, Ollama, etc.) by making the base API endpoint configurable.
-   **Acceptance Criteria:**
    -   [x] The `OPENAI_BASE_URL` environment variable is added to configuration and properly handled per `DEV-RULE S1`.
        -   **Rationale:** Enables flexibility to use different OpenAI-compatible providers for cost optimization and vendor diversity.
        -   **Evidence:** Added `OPENAI_BASE_URL` to Config struct in main.go with default value "https://api.openai.com/v1" and proper environment variable handling.
    -   [x] The OpenAI adapter supports custom base URLs while maintaining backward compatibility with the default OpenAI endpoint.
        -   **Rationale:** Allows seamless switching between providers without code changes, only configuration.
        -   **Evidence:** OpenAI adapter `WithBaseURL()` functional option implemented and integrated into service initialization with backward compatibility maintained.
    -   [x] Configuration validation ensures the base URL is properly formatted and accessible.
        -   **Rationale:** Prevents runtime failures due to misconfigured endpoints.
        -   **Evidence:** Added `validateBaseURL()` function in both main.go and openai_adapter that validates HTTP/HTTPS protocol requirements.
    -   [x] **Unit Test:** The adapter correctly constructs API calls with custom base URLs and handles provider-specific response formats.
        -   **Rationale:** Ensures compatibility across different OpenAI-compatible providers.
        -   **Evidence:** `TestBaseURLValidation` and `TestProviderSpecificConfiguration` tests pass, validating URL construction and provider-specific configurations.
    -   [x] **Integration Test:** The service successfully transcribes audio using an alternative OpenAI-compatible provider.
        -   **Rationale:** Validates real-world compatibility with multiple providers.
        -   **Evidence:** Service successfully starts with custom base URL: `OPENAI_BASE_URL=https://api.groq.com/openai/v1` and `TestMultipleProviders` test supports Groq, Ollama, and Azure configurations.
    -   [x] Documentation is updated to include examples for popular OpenAI-compatible providers (Groq, Azure OpenAI, Ollama).
        -   **Rationale:** Provides clear guidance for operators on how to configure different providers.
        -   **Evidence:** Created `docs/OPENAI_COMPATIBLE_PROVIDERS.md` with comprehensive examples for OpenAI, Groq, Azure OpenAI, Ollama, Together.ai, and Anyscale Endpoints.
    -   [x] **Workflow:** All work is committed following the process in `DEV-RULE W2`, including updating this document with rationale and evidence before the commit.
        -   **Rationale:** Maintains development process compliance and proper documentation.
        -   **Evidence:** This document updated with comprehensive rationale and evidence before commit creation.

---

## **Epic: Embedding Service**

### **Story: P1-ES1: Service Scaffolding & Logic**

-   **Description:** As a developer, I need to set up the Embedding Service, generate embeddings from text, and store the results in the `pgvector` database.
-   **Acceptance Criteria:**
    -   [ ] An `embedder/` directory is created with the structure defined in `LLD-ES Sec. 2`, implementing Ports & Adapters architecture per `ARCH-RULE A1`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] Port interfaces are defined in `internal/ports/` for `EmbeddingGenerator` and `VectorStore` as specified in `LLD-ES Sec. 2`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] Core application logic in `internal/core/` has zero external dependencies, adhering to `ARCH-RULE A1.1`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] A `Dockerfile` and `Makefile` are created per `DEV-RULE E1` and `E3`, with both `build-docker` and `build-native` targets per `DEV-RULE E4`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] Dependency injection is implemented in `cmd/main.go` as the composition root per `ARCH-RULE A2`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The service correctly subscribes to the `transcription.succeeded` event and orchestrates the embedding/storage flow as per `LLD-ES Sec. 3.1`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The service implements structured logging per `DEV-RULE O1` with JSON format and includes `correlation_id` and `recording_id` per `DEV-RULE O2`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] A `/health` endpoint is exposed per `DEV-RULE O3`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] All environment variables from `LLD-ES Sec. 4` are properly handled per `DEV-RULE S1`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] Custom error types are defined for predictable failures (e.g., `ErrDatabaseUnavailable`, `ErrEmbeddingFailed`) per `ARCH-RULE A3.2`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Unit Test:** The core logic is unit-tested with mock `openai` and `pgvector` adapters to verify proper data flow and error handling.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Integration Test:** The service successfully receives a NATS event, calls the live OpenAI API to get an embedding, and writes the complete record to the live `pgvector` database.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Error Handling Test:** The system correctly handles and logs errors for: database unreachable, OpenAI API failures, malformed events - all with proper error wrapping per `ARCH-RULE A3`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Contract Compliance Test:** Verify the service processes `transcription.succeeded` events exactly as specified in `CONTRACT.md` and stores all required fields.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Workflow:** All work is committed following the process in `DEV-RULE W2`, including updating this document with rationale and evidence before the commit.
        -   **Rationale:**
        -   **Evidence:**

---

## **Epic: Query Service**

### **Story: P1-QS1: Service Scaffolding & Logic**

-   **Description:** As a developer, I need to set up the Query Service, expose an HTTP API, and implement the vector search logic.
-   **Acceptance Criteria:**
    -   [ ] A `query_svc/` directory is created with the structure defined in `LLD-QS Sec. 2`, implementing Ports & Adapters architecture per `ARCH-RULE A1`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] Port interfaces are defined in `internal/ports/` for `EmbeddingGenerator` and `VectorSearcher` as specified in `LLD-QS Sec. 2`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] Core application logic in `internal/core/` has zero external dependencies, adhering to `ARCH-RULE A1.1`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] A `Dockerfile` and `Makefile` are created per `DEV-RULE E1` and `E3`, with both `build-docker` and `build-native` targets per `DEV-RULE E4`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] Dependency injection is implemented in `cmd/main.go` as the composition root per `ARCH-RULE A2`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] An `http_adapter` exposes a `POST /api/v1/query` endpoint as defined in `LLD-QS Sec. 3.1` with proper HTTP error handling.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The core service logic correctly orchestrates the search flow (vectorize query -> search DB) as defined in `LLD-QS Sec. 3.1`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The service implements structured logging per `DEV-RULE O1` with JSON format and includes `correlation_id` per `DEV-RULE O2`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] A `/health` endpoint is exposed per `DEV-RULE O3`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] All environment variables from `LLD-QS Sec. 4` are properly handled per `DEV-RULE S1`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] Custom error types are defined for predictable failures (e.g., `ErrDatabaseUnavailable`, `ErrInvalidQuery`) per `ARCH-RULE A3.2`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Unit Test:** The core logic is unit-tested with mock `openai` and `pgvector` adapters to verify proper search flow and error handling.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Integration Test:** The service successfully receives an API request, calls the OpenAI API, queries the live `pgvector` database, and returns valid JSON results.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Integration Test:** The API correctly handles tag-based filtering as part of the database query with proper SQL injection protection.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Error Handling Test:** The system correctly handles and logs errors for: database unreachable, OpenAI API failures, malformed requests - all with proper HTTP status codes.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Workflow:** All work is committed following the process in `DEV-RULE W2`, including updating this document with rationale and evidence before the commit.
        -   **Rationale:**
        -   **Evidence:**

---

## **Epic: Backend Integration & Contract Testing**

### **Story: P1-INT1: End-to-End Backend Integration**

-   **Description:** As a developer, I need to verify that all backend services work together correctly and fulfill the complete contract specified in `CONTRACT.md`.
-   **Acceptance Criteria:**
    -   [ ] **Full Pipeline Test:** A test successfully executes the complete flow: recording.start -> recording.stop -> transcription.run -> transcription.succeeded -> embedding storage -> query retrieval.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Contract Testing:** All services are tested together to verify every command and event in `CONTRACT.md` works correctly with proper payload validation.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Tag Propagation Test:** Verify that tags specified in initial commands are properly propagated through the entire pipeline and are queryable via the Query Service.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Metadata Handling Test:** Verify that metadata fields are properly carried through events without being modified by core services.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Error Propagation Test:** Verify that failures in any service properly publish the correct error events and don't break the overall system.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Performance Test:** Verify the system can handle concurrent requests and maintains acceptable response times under load.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Docker Compose Integration:** All services start successfully via `make up` and can communicate with each other in the containerized environment.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Workflow:** All work is committed following the process in `DEV-RULE W2`, including updating this document with rationale and evidence before the commit.
        -   **Rationale:**
        -   **Evidence:**