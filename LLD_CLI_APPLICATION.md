# LLD: The CLI Application

**ID:** LLD-CLI
**Version:** 1.0
**Status:** Proposed

---

### 1. Overview

This document provides the low-level design for the **`speakr` CLI Application**, the primary user-facing tool.

-   **Purpose:** To provide a simple, intuitive command-line interface for users to record audio, have it transcribed, and get the result copied to their clipboard.
-   **Architectural Adherence:** The CLI acts as a "driving adapter" for the entire backend system. It **must** adhere to the `CONTRACT.md` for all its interactions with the NATS bus.

### 2. Core Components & Packages

The application will be structured into the following Go packages within the `cli/` directory:

-   `cmd/`: The main entry point for the application, using a library like `cobra` to manage commands and flags.
-   `internal/app`: The core application logic for the CLI, responsible for orchestrating the user workflow.
-   `internal/nats`: A dedicated package to handle all communication with the NATS server (publishing commands, subscribing to events).
-   `internal/recorder`: A wrapper around the OS-specific `ffmpeg` commands.
-   `internal/clipboard`: A wrapper around the OS-specific clipboard commands (`pbcopy`, `xclip`, `clip`).

### 3. Logic Flow

#### 3.1. Default `speakr` command execution

1.  The `main` function in `cmd/` initializes the application.
2.  It establishes a connection to the NATS server via the `nats` package.
3.  It publishes a `speakr.command.recording.start` command to NATS.
4.  It simultaneously subscribes to the `speakr.event.recording.started` and `speakr.event.transcription.succeeded` topics, filtering by a unique `correlation_id` it generated for this session.
5.  **On `recording.started` event:**
    -   It receives the `recording_id`.
    -   It prints "🔴 Recording... Press Enter to stop." to the console.
    -   It uses the `recorder` package to start `ffmpeg`.
6.  The application waits for the user to press the `Enter` key.
7.  Upon `Enter`, it publishes a `speakr.command.recording.stop` command, including the `recording_id` and setting `metadata.copy_to_clipboard` to `true`.
8.  It prints "✅ Recording finished. Transcribing..."
9.  **On `transcription.succeeded` event:**
    -   It receives the event.
    -   It inspects the `metadata` field. If `copy_to_clipboard` is `true`, it uses the `clipboard` package to copy the `transcribed_text`.
    -   It prints "✔ Transcribed text copied to clipboard!"
    -   The application exits.

### 4. Configuration (Environment Variables)

-   `NATS_URL`: URL for the NATS server.
-   `OPENAI_API_KEY`: (Potentially needed if any direct-to-API functionality is added later, but not for the core NATS-driven flow).
