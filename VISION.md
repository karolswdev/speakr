# Project Vision: Speakr

This document outlines the strategic vision for the `speakr` project, from its simplest application to its potential as a core component in a large-scale, event-driven architecture.

---

### The Core Idea: Frictionless Voice-to-Text

At its heart, `speakr` is designed to be the fastest, most frictionless way for a user to get spoken words into a usable text format. The initial and primary use case is a simple, powerful CLI tool that removes all barriers between thought and text.

**Use Case 1: The Keyboard Macro**

A user configures a global keyboard shortcut (e.g., `Ctrl+Shift+Space`). Pressing this shortcut invokes the `speakr` CLI.

1.  The `speakr` tool immediately starts recording audio from the default microphone.
2.  The user speaks their thoughts.
3.  The user presses `Enter` or a second shortcut to stop.
4.  `speakr` transcribes the audio and automatically places the text into the system clipboard.

The entire process is seamless. The user never has to leave their current application. It's a "fire and forget" tool for capturing notes, writing emails, or drafting code.

### The Broader Vision: A Decoupled Voice Services Platform

While the CLI is the entry point, the architecture of `speakr` is intentionally designed for a much larger purpose. By building the application around a formal, message-based contract and the Ports & Adapters pattern, `speakr` evolves from a simple tool into a robust, backend **Voice Services Platform**.

**Use Case 2: An Event-Driven Microservice**

In this vision, the `speakr` service runs as a headless component within a larger data pipeline, orchestrated by a message bus like NATS.

-   **Input from Anywhere:** Other services can now use `speakr`'s capabilities. For example, a chat application could save a voice message, publish a `speakr.command.transcription.run` command with the audio data, and listen for the `transcription.succeeded` event to display the transcript within the chat UI.
-   **Data Traceability:** By leveraging a system of user-defined `tags`, all data flowing through `speakr` becomes traceable and queryable. A project management dashboard could be built to search and display all transcripts related to a specific project (`"tag": "project-phoenix"`), regardless of their origin.
-   **Extensible Capabilities:** The contract-driven design means new capabilities can be added without modifying the core. The platform can grow to include:
    -   **Speaker Diarization:** Identifying who spoke and when.
    -   **Summarization:** Using an LLM to create summaries of long transcripts.
    -   **Sentiment Analysis:** Determining the emotional tone of the text.
    -   **Translation:** Translating the final transcript into other languages.

`speakr` is not just a tool; it's a foundational building block for any system that needs to process, understand, or react to human speech. It is designed to be simple enough for a single user's desktop but powerful enough to serve an entire enterprise.
