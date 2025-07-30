# Development Plan: Phase 0 - Foundation & Environment Setup

**Version:** 1.0
**Status:** Ready for Development

---

### **Objective**

To create a stable, documented, and automated foundation for all future development. This phase ensures that when development teams start Phase 1 and 2, they have everything they need to be productive immediately.

### **Referenced Documents**

-   **Vision & Requirements:**
    -   `VISION.md`
-   **Rules & Architecture:**
    -   `ARCHITECTURE.md`
    -   `DEVELOPMENT_RULES.md`
-   **Contracts & Interfaces:**
    -   `CONTRACT.md`
-   **Low-Level Designs:**
    -   `LLD_TRANSCRIBER_SERVICE.md`
    -   `LLD_EMBEDDING_SERVICE.md`
    -   `LLD_QUERY_SERVICE.md`
    -   `LLD_CLI_APPLICATION.md`

---

## **Stories**

### **Story: P0-S1: Initialize Project Structure & Version Control**

-   **Epic:** Foundation
-   **Description:** As a developer, I need the basic project scaffolding and version control setup so I can start working on the codebase.
-   **Acceptance Criteria:**
    -   [ ] A `git` repository is initialized.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] A `.gitignore` file suitable for Go projects is created and committed, including `.env` files per `DEV-RULE S2`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The `go.mod` file is created using `go mod init speakr`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The initial directory structure is created following the LLD specifications: `transcriber/`, `embedder/`, `query_svc/`, `cli/`, `scripts/`, `.github/workflows/`, `terraform/`, `docs/`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] Each service directory contains the package structure defined in their respective LLD documents (e.g., `cmd/`, `internal/core/`, `internal/adapters/`, `internal/ports/`).
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] A `.env.example` file is created showing all required environment variables from all LLD documents.
        -   **Rationale:**
        -   **Evidence:**

### **Story: P0-S2: Implement Docker Compose Environment**

-   **Epic:** Foundation
-   **Description:** As a developer, I need a `docker-compose.yml` file to easily spin up all the necessary infrastructure services for local development per `DEV-RULE E2`.
-   **Acceptance Criteria:**
    -   [ ] A `docker-compose.yml` file exists at the project root adhering to `DEV-RULE E2.1` (one-command startup).
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The file defines services for `nats`, `minio`, and `postgres` with proper networking configuration.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The `postgres` service uses an image that includes the `pgvector` extension (e.g., `pgvector/pgvector:pg16`) as required by the Embedding and Query services.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The services are configured to expose their default ports to the host machine (NATS: 4222, MinIO: 9000/9001, Postgres: 5432) to support `DEV-RULE E4.2` native builds.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] Persistent volumes are configured for `minio` and `postgres` to retain data between restarts.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] Environment variables are properly configured for each service, including MinIO credentials and Postgres database settings.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The `docker-compose up` command successfully starts all services without errors and all services report healthy status.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Integration Test:** Verify connectivity to each service from the host machine using appropriate client tools (nats-cli, mc, psql).
        -   **Rationale:**
        -   **Evidence:**

### **Story: P0-S3: Implement Infrastructure Seeding Scripts**

-   **Epic:** Foundation
-   **Description:** As a developer, I need automated scripts to initialize the infrastructure services to a known state on startup per `DEV-RULE E2.2`.
-   **Acceptance Criteria:**
    -   [ ] A `scripts/seed/` directory exists with proper organization for each service type.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] A SQL script (`init.sql`) exists that creates the necessary database schema for the Embedding Service, including the `pgvector` extension and tables with proper indexing for vector similarity search.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The database schema includes all fields specified in `LLD-ES` and `LLD-QS`: `recording_id`, `transcribed_text`, `tags`, `embedding` vector, and proper constraints.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] A script exists to create the required MinIO bucket (`speakr-audio`) and configure proper access policies as specified in `LLD-TS`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] A script exists to configure NATS streams and subjects according to the `CONTRACT.md` specifications.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The `docker-compose.yml` is configured to automatically run these seeding scripts on the first startup of the relevant services, ensuring idempotent execution.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Verification Test:** After running `make up`, all infrastructure is properly seeded and ready for service connections without manual intervention.
        -   **Rationale:**
        -   **Evidence:**

### **Story: P0-S4: Create Root Makefile**

-   **Epic:** Foundation
-   **Description:** As a developer, I need a root `Makefile` to provide a consistent interface for common project tasks per `DEV-RULE E3`.
-   **Acceptance Criteria:**
    -   [ ] A `Makefile` exists at the project root with standardized targets as required by `DEV-RULE E3`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The `Makefile` contains an `up` target that runs `docker-compose up -d` and waits for all services to be healthy.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The `Makefile` contains a `down` target that runs `docker-compose down` and cleans up properly.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The `Makefile` contains a `logs` target to follow the logs of all services with proper formatting.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The `Makefile` contains a `clean` target to stop services and remove volumes, ensuring complete environment reset.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The `Makefile` contains targets for `build-docker`, `build-native`, `test`, `lint` that delegate to individual service Makefiles.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The `Makefile` includes a `health-check` target that verifies all infrastructure services are responding correctly.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Verification Test:** All Makefile targets execute successfully and provide clear, actionable output.
        -   **Rationale:**
        -   **Evidence:**

### **Story: P0-S5: Setup Basic CI/CD Pipeline**

-   **Epic:** Foundation
-   **Description:** As a developer, I need a basic CI pipeline in GitHub Actions to ensure code quality and consistency per `DEV-RULE C1` and `C2`.
-   **Acceptance Criteria:**
    -   [ ] A `.github/workflows/ci.yml` file exists implementing the requirements from `DEV-RULE C1`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The workflow is triggered on every push to a feature branch and on every pull request to `main` as specified in `DEV-RULE C2`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The workflow includes a job for linting the Go code using `golangci-lint` with strict configuration.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The workflow includes a job for running Go unit tests (`go test ./...`) with coverage reporting.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The workflow includes a job for building Docker containers for each service to verify `DEV-RULE E1` compliance.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The workflow includes validation that all services can start successfully in the CI environment.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The workflow enforces that all status checks must pass before PR merge, implementing `DEV-RULE Q2`.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Integration Test:** The CI pipeline successfully runs on a test PR and blocks merge when tests fail.
        -   **Rationale:**
        -   **Evidence:**
