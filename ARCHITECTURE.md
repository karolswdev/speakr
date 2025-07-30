# Speakr Architectural Principles

This document outlines the authoritative set of architectural rules and patterns to be followed during the development of the `speakr` project. Adherence to these principles is mandatory to ensure the creation of a robust, maintainable, and extensible system.

---

### A1: Ports & Adapters (Hexagonal Architecture)

The project **must** adhere to the Ports & Adapters architecture. This is the foundational principle that governs the overall structure of the application.

-   **A1.1: Core Logic Isolation:** The core application logic (the "hexagon") must be completely isolated from any external infrastructure or implementation details. The core must not have any knowledge of specific technologies like `ffmpeg`, `NATS`, or the `OpenAI API`.
-   **A1.2: Ports as Interfaces:** The core logic must define its dependencies on the outside world through `ports`, which in Go are defined as `interface` types. These interfaces represent the contracts the core needs to function (e.g., `AudioRecorder`, `TranscriptionService`, `EventPublisher`).
-   **A1.3: Adapters as Implementations:** Concrete implementations of the ports are called `adapters`. These adapters are responsible for interacting with external technologies (e.g., an `FFmpegAdapter` that implements the `AudioRecorder` interface). Adapters live outside the core logic.

### A2: Dependency Injection (DI)

All dependencies required by the core logic **must** be provided from the outside using Dependency Injection.

-   **A2.1: No Self-Instantiation:** The core logic must not instantiate its own dependencies. For example, the core application service must not call `NewFFmpegAdapter()`.
-   **A2.2: Composition Root:** The `main.go` file (or a dedicated factory/builder) acts as the "composition root." It is responsible for creating the concrete adapter instances and "injecting" them into the core application service during its construction. This is the only place where the core logic and the concrete implementations are wired together.

### A3: Explicit Error Handling

The application **must** employ a strategy of explicit and contextual error handling.

-   **A3.1: Error Wrapping:** Errors returned from adapters or other low-level functions **must** be wrapped with additional context. Use `fmt.Errorf("operation failed: %w", err)` to create a chain of errors that provides a clear trace of the failure.
-   **A3.2: Custom Error Types:** For specific, predictable error conditions that require programmatic handling (e.g., an API key not being set), custom error variables **must** be defined (e.g., `var ErrAPIKeyNotSet = errors.New(...)`). This allows for reliable error checking using `errors.Is()`.

### A4: Graceful Shutdown

The application **must** handle process interruptions gracefully to ensure resources are not leaked.

-   **A4.1: Context Propagation:** Go's `context.Context` **must** be used to manage the lifecycle of long-running operations, such as recording audio or making API calls.
-   **A4.2: Signal Handling:** The application's entry point **must** listen for OS interrupt signals (`SIGINT`, `SIGTERM`). Upon receiving a signal, the root context must be canceled, allowing all downstream operations to terminate cleanly and perform necessary cleanup (e.g., deleting temporary files).

### A5: Component Configuration

Configuration for components (adapters) **must** be provided in a flexible and extensible manner.

-   **A5.1: Functional Options Pattern:** For components that require more than one or two configuration parameters, the Functional Options Pattern **must** be used. This is the idiomatic Go approach for creating clean, readable, and forward-compatible constructors.

### A6: Contract-Driven Development

The application's functionality **must** be exposed via a formal, technology-agnostic contract.

-   **A6.1: Commands and Events:** The contract is defined by the **Commands** the system can accept and the **Events** it can emit. This makes the system's capabilities explicit and easy to understand. (See `CONTRACT.md` for the full definition).
