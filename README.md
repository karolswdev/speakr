# Speakr

A simple, powerful CLI tool and voice services platform to record, transcribe, and manage audio.

This project is built following the principles outlined in `ARCHITECTURE.md` and `VISION.md`.
The service's capabilities are formally defined in `CONTRACT.md`.

## Prerequisites

-   Go (1.21 or newer)
-   `ffmpeg`
-   `xclip` (on Linux)

## Getting Started

1.  **Set your OpenAI API Key:**
    ```sh
    export OPENAI_API_KEY='your-key-here'
    ```

2.  **Build the binary:**
    ```sh
    make build
    ```

3.  **Run the application:**
    ```sh
    ./speakr
    ```
