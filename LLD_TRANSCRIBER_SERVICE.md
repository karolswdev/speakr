# LLD: The Transcriber Service

**ID:** LLD-TS
**Version:** 1.0
**Status:** Proposed

---

### 1. Overview

This document provides the low-level design for the **Transcriber Service**, the core backend component of the `speakr` ecosystem.

-   **Purpose:** To accept audio data via NATS commands, orchestrate the recording and transcription process, and publish the results as events.
-   **Architectural Adherence:** This service **must** strictly adhere to all principles outlined in `ARCHITECTURE.md`.

### 2. Core Components & Packages

The service will be structured into the following Go packages within the `transcriber/` directory:

-   `cmd/`: The main entry point for the service. Responsible for the composition root (wiring dependencies) and starting the service.
-   `internal/core`: The implementation of the core application logic (the "hexagon"). It is pure and has no knowledge of external infrastructure.
-   `internal/adapters`: Contains all concrete implementations of the ports.
    -   `nats_adapter/`: Implements the NATS subscriber and publisher.
    -   `ffmpeg_adapter/`: Implements the `AudioRecorder` port.
    -   `openai_adapter/`: Implements the `TranscriptionService` port.
    -   `minio_adapter/`: Implements the `ObjectStore` port for saving audio files.
-   `internal/ports`: Defines the Go interfaces for all dependencies required by the core logic (e.g., `AudioRecorder`, `TranscriptionService`, `ObjectStore`, `EventPublisher`).

### 3. Logic Flow

#### 3.1. On `speakr.command.recording.start`

1.  The `nats_adapter` receives the command.
2.  It calls the `core` service's `StartRecording` method.
3.  The `core` service generates a `recording_id` and `tags`.
4.  It invokes the `ffmpeg_adapter` to start a new recording process.
5.  Upon successful start, it uses the `nats_adapter` to publish a `speakr.event.recording.started` event with the `recording_id` and `tags`.

#### 3.2. On `speakr.command.recording.stop`

1.  The `nats_adapter` receives the command.
2.  It calls the `core` service's `StopRecording` method.
3.  The `core` service commands the `ffmpeg_adapter` to stop the recording process identified by `recording_id`.
4.  The `ffmpeg_adapter` finalizes the audio file.
5.  The `core` service commands the `minio_adapter` to upload the audio file to object storage, using the `recording_id` as the object key.
6.  The `core` service publishes a `speakr.event.recording.finished` event.
7.  If `transcribe_on_stop` is `true`, it immediately issues a `speakr.command.transcription.run` command for the same `recording_id`.

#### 3.3. On `speakr.command.transcription.run`

1.  The `nats_adapter` receives the command.
2.  It calls the `core` service's `TranscribeAudio` method.
3.  The `core` service retrieves the audio file from the `minio_adapter` using the `recording_id`.
4.  It passes the audio data to the `openai_adapter`.
5.  The `openai_adapter` sends the data to the OpenAI Whisper API.
6.  Upon receiving the transcript, the `core` service publishes a `speakr.event.transcription.succeeded` event containing the text and tags.

### 4. Configuration (Environment Variables)

-   `NATS_URL`: URL for the NATS server.
-   `OPENAI_API_KEY`: API key for the OpenAI service.
-   `MINIO_ENDPOINT`: Endpoint URL for the MinIO server.
-   `MINIO_ACCESS_KEY`: Access key for MinIO.
-   `MINIO_SECRET_KEY`: Secret key for MinIO.
-   `MINIO_BUCKET_NAME`: Name of the bucket to store audio files.
