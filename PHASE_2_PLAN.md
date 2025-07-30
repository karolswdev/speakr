# Development Plan: Phase 2 - CLI Application Development

**Version:** 1.0
**Status:** Ready for Development

---

### **Objective**

To build, test, and deliver the primary user-facing `speakr` CLI application. This application will act as a "driving adapter" for the backend platform, providing a polished and intuitive user experience for recording, transcription, and clipboard management.

### **Referenced Documents**

-   **Vision & Requirements:** `VISION.md`
-   **Rules & Architecture:** `ARCHITECTURE.md`, `DEVELOPMENT_RULES.md`
-   **Contracts & Interfaces:** `CONTRACT.md`
-   **Low-Level Designs:** `LLD_CLI_APPLICATION.md`

---

## **Epic: Core Application & NATS Integration**

### **Story: P2-CLI1: Service Scaffolding & NATS Client**

-   **Description:** As a developer, I need to set up the initial CLI application structure and implement a robust NATS client to communicate with the backend.
-   **Acceptance Criteria:**
    -   [ ] A `cli/` directory is created with the package structure defined in `LLD-CLI Sec. 2`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] A `Dockerfile` and `Makefile` are created, adhering to `DEV-RULE E1` and `E3`. The `Makefile` **must** include the `build-native` target as per `DEV-RULE E4.2`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] A NATS client is implemented that can publish commands and subscribe to events as defined in `CONTRACT.md`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Unit Test:** The NATS client logic is unit-tested using a mock NATS connection.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Integration Test:** The CLI application, when run natively, successfully connects to the NATS server from the Docker Compose environment.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Error Handling Test:** The application exits gracefully with a clear error message if it cannot connect to the NATS server.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Workflow:** All work is committed following the process in `DEV-RULE W2`, including updating this document with rationale and evidence before the commit.
        -   **Rationale:**
        -   **Evidence:**

### **Story: P2-CLI2: Implement Core Recording Workflow**

-   **Description:** As a developer, I need to implement the main application workflow that orchestrates the recording and transcription process over NATS.
-   **Acceptance Criteria:**
    -   [ ] The core application logic implements the stateful workflow described in `LLD-CLI Sec. 3.1`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The CLI correctly publishes `recording.start` and `recording.stop` commands.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The CLI correctly subscribes to and handles the `recording.started` and `transcription.succeeded` events.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Unit Test:** The entire workflow is unit-tested using a mock NATS client to simulate the back-and-forth event flow.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Integration Test:** The CLI successfully completes the full recording-to-transcription loop by communicating with a live (mocked or real) backend over NATS.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Workflow:** All work is committed following the process in `DEV-RULE W2`, including updating this document with rationale and evidence before the commit.
        -   **Rationale:**
        -   **Evidence:**

---

## **Epic: Hardware & OS Integration**

### **Story: P2-CLI3: Implement `ffmpeg` Recorder Adapter**

-   **Description:** As a developer, I need to create a hardware adapter that can start and stop `ffmpeg` to record audio from the system microphone.
-   **Acceptance Criteria:**
    -   [ ] A `recorder` package is created that wraps the execution of `ffmpeg` as a subprocess, as per `LLD-CLI Sec. 2`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The implementation includes OS-specific commands for Linux, macOS, and Windows.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Integration Test:** The `build-native` binary is executed on a host machine and successfully records a `.wav` file using the installed `ffmpeg`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Error Handling Test:** The application exits gracefully with a clear error message if `ffmpeg` is not found in the system's PATH.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Workflow:** All work is committed following the process in `DEV-RULE W2`, including updating this document with rationale and evidence before the commit.
        -   **Rationale:**
        -   **Evidence:**

### **Story: P2-CLI4: Implement Clipboard Adapter**

-   **Description:** As a developer, I need to create an OS adapter that can copy the final transcript to the system clipboard.
-   **Acceptance Criteria:**
    -   [ ] A `clipboard` package is created that wraps the execution of `pbcopy`, `xclip`, and `clip`, as per `LLD-CLI Sec. 2`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Integration Test:** The `build-native` binary is executed, and upon receiving a test transcript, successfully copies it to the host OS clipboard.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Error Handling Test:** On Linux, the application provides a clear error message if neither `xclip` nor `xsel` is available.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Workflow:** All work is committed following the process in `DEV-RULE W2`, including updating this document with rationale and evidence before the commit.
        -   **Rationale:**
        -   **Evidence:**

---

## **Epic: User Experience**

### **Story: P2-CLI5: Refine UX and Implement Command-Line Flags**

-   **Description:** As a developer, I need to polish the user experience by adding clear status messages and command-line flags for controlling behavior.
-   **Acceptance Criteria:**
    -   [ ] A CLI library like `cobra` is integrated to manage commands and flags.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The CLI prints clear, user-friendly status messages for each step of the process (e.g., "Connecting...", "🔴 Recording...", "✅ Transcription complete!").
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] A `--no-clipboard` flag is implemented that prevents the final text from being copied to the clipboard.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] A `--tags` flag is implemented that accepts a comma-separated list of strings and includes them in the `recording.start` command payload.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Unit Test:** The flag parsing and conditional logic are unit-tested.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Workflow:** All work is committed following the process in `DEV-RULE W2`, including updating this document with rationale and evidence before the commit.
        -   **Rationale:**
        -   **Evidence:**
