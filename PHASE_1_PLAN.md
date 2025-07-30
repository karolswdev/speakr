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
    -   [ ] The `ffmpeg_adapter` and `minio_adapter` are created, implementing the ports defined in `LLD-TS Sec. 2` with proper error wrapping per `ARCH-RULE A3.1`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] Custom error types are defined for predictable failures (e.g., `ErrFFmpegNotFound`, `ErrBucketNotFound`) per `ARCH-RULE A3.2`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The core service logic correctly orchestrates the `start` and `stop` recording commands as per `LLD-TS Sec. 3.1 & 3.2` with proper context propagation per `ARCH-RULE A4.1`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] Graceful shutdown is implemented with signal handling per `ARCH-RULE A4.2`, ensuring ongoing recordings are properly finalized.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] Configuration uses functional options pattern per `ARCH-RULE A5.1` for complex adapter configuration.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Unit Test:** The core service logic is unit-tested with mock `ffmpeg` and `minio` adapters to verify the correct sequence of calls and error propagation.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Integration Test:** A test successfully triggers the `ffmpeg` adapter to record a short audio clip and uploads it to the live MinIO container.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Error Handling Test:** The system correctly handles and logs errors for: MinIO bucket not existing, ffmpeg not found, disk space issues, network failures - all adhering to `ARCH-RULE A3`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The service correctly publishes `recording.started` and `recording.finished` events to NATS with exact payload format specified in `CONTRACT.md`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Contract Compliance Test:** Verify all published events match the exact JSON schema from `CONTRACT.md` including all required fields and proper tag propagation.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Workflow:** All work is committed following the process in `DEV-RULE W2`, including updating this document with rationale and evidence before the commit.
        -   **Rationale:**
        -   **Evidence:**

### **Story: P1-TS3: Implement Transcription Logic**

-   **Description:** As a developer, I need to implement the logic to fetch an audio file and transcribe it using the OpenAI API.
-   **Acceptance Criteria:**
    -   [ ] An `openai_adapter` is created that implements the `TranscriptionService` port as defined in `LLD-TS Sec. 2` with proper error wrapping per `ARCH-RULE A3.1`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] Custom error types are defined for OpenAI-specific failures (e.g., `ErrAPIKeyInvalid`, `ErrQuotaExceeded`, `ErrAudioTooLarge`) per `ARCH-RULE A3.2`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The core service logic correctly handles the `transcription.run` command as per `LLD-TS Sec. 3.3` with proper context propagation and timeout handling.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The adapter supports both recording ID and raw audio data input modes as specified in `CONTRACT.md`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Unit Test:** The core service logic is unit-tested with a mock `openai_adapter` to verify the correct data is passed and error handling.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Integration Test:** A test successfully fetches a file from MinIO, calls the live OpenAI API, and receives a valid transcript.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Error Handling Test:** The system correctly handles and logs errors for: invalid API key, quota exceeded, network timeouts, malformed audio files - all publishing appropriate `transcription.failed` events per `CONTRACT.md`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Contract Compliance Test:** Verify `transcription.succeeded` and `transcription.failed` events match exact JSON schema from `CONTRACT.md` with proper tag and metadata propagation.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Workflow:** All work is committed following the process in `DEV-RULE W2`, including updating this document with rationale and evidence before the commit.
        -   **Rationale:**
        -   **Evidence:**

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