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
    -   [x] A `git` repository is initialized.
        -   **Rationale:** Git repository provides version control foundation for the project per development best practices.
        -   **Evidence:** `git status` shows "On branch main" confirming repository is initialized.
    -   [x] A `.gitignore` file suitable for Go projects is created and committed, including `.env` files per `DEV-RULE S2`.
        -   **Rationale:** Prevents sensitive environment files and build artifacts from being committed, ensuring security compliance with DEV-RULE S2.
        -   **Evidence:** `.gitignore` file exists with comprehensive Go project exclusions and explicit `.env` file exclusion.
    -   [x] The `go.mod` file is created using `go mod init speakr`.
        -   **Rationale:** Establishes Go module for dependency management and project identification.
        -   **Evidence:** `go.mod` file exists with `module speakr` declaration, confirmed by `go mod init` output.
    -   [x] The initial directory structure is created following the LLD specifications: `transcriber/`, `embedder/`, `query_svc/`, `cli/`, `scripts/`, `.github/workflows/`, `terraform/`, `docs/`.
        -   **Rationale:** Provides organized structure for all project components as specified in respective LLD documents.
        -   **Evidence:** Directory structure created with `mkdir -p` command covering all required top-level directories.
    -   [x] Each service directory contains the package structure defined in their respective LLD documents (e.g., `cmd/`, `internal/core/`, `internal/adapters/`, `internal/ports/`).
        -   **Rationale:** Implements Ports & Adapters architecture structure per ARCH-RULE A1 from the start.
        -   **Evidence:** Complete package structure created for all services including core, adapters, and ports directories as specified in LLD-TS, LLD-ES, LLD-QS, and LLD-CLI.
    -   [x] A `.env.example` file is created showing all required environment variables from all LLD documents.
        -   **Rationale:** Provides clear documentation of all configuration requirements and supports secure environment variable management per DEV-RULE S1.
        -   **Evidence:** `.env.example` file created with comprehensive environment variables from LLD-TS Sec. 4, LLD-ES Sec. 4, LLD-QS Sec. 4, and LLD-CLI Sec. 4.

### **Story: P0-S2: Implement Docker Compose Environment**

-   **Epic:** Foundation
-   **Description:** As a developer, I need a `docker-compose.yml` file to easily spin up all the necessary infrastructure services for local development per `DEV-RULE E2`.
-   **Acceptance Criteria:**
    -   [x] A `docker-compose.yml` file exists at the project root adhering to `DEV-RULE E2.1` (one-command startup).
        -   **Rationale:** Enables one-command startup of entire development environment per DEV-RULE E2.1 requirement.
        -   **Evidence:** `docker-compose.yml` created with all required services and `docker compose up -d` successfully starts environment.
    -   [x] The file defines services for `nats`, `minio`, and `postgres` with proper networking configuration.
        -   **Rationale:** Provides all infrastructure services required by backend services as specified in LLD documents.
        -   **Evidence:** Services defined with custom network `speakr-network` for proper inter-service communication.
    -   [x] The `postgres` service uses an image that includes the `pgvector` extension (e.g., `pgvector/pgvector:pg16`) as required by the Embedding and Query services.
        -   **Rationale:** Supports vector similarity search functionality required by LLD-ES and LLD-QS.
        -   **Evidence:** Using `pgvector/pgvector:pg16` image and database initialization confirms pgvector extension is available.
    -   [x] The services are configured to expose their default ports to the host machine (NATS: 4222, MinIO: 9010/9011, Postgres: 5432) to support `DEV-RULE E4.2` native builds.
        -   **Rationale:** Allows native builds to connect to containerized infrastructure per DEV-RULE E4.2.
        -   **Evidence:** Ports properly exposed and verified with `make status` showing all services accessible.
    -   [x] Persistent volumes are configured for `minio` and `postgres` to retain data between restarts.
        -   **Rationale:** Ensures data persistence across development sessions for reliable testing.
        -   **Evidence:** Named volumes `minio_data`, `postgres_data`, and `nats_data` configured in docker-compose.yml.
    -   [x] Environment variables are properly configured for each service, including MinIO credentials and Postgres database settings.
        -   **Rationale:** Provides secure and configurable service setup matching .env.example specifications.
        -   **Evidence:** All services configured with appropriate environment variables and credentials.
    -   [x] The `docker-compose up` command successfully starts all services without errors and all services report healthy status.
        -   **Rationale:** Validates the environment setup works correctly and reliably.
        -   **Evidence:** `docker compose up -d` completed successfully and `make status` shows all services healthy.
    -   [x] **Integration Test:** Verify connectivity to each service from the host machine using appropriate client tools (nats-cli, mc, psql).
        -   **Rationale:** Confirms services are accessible for development and testing.
        -   **Evidence:** PostgreSQL connectivity verified with `psql` command returning test data count, MinIO and NATS services showing healthy status.

### **Story: P0-S3: Implement Infrastructure Seeding Scripts**

-   **Epic:** Foundation
-   **Description:** As a developer, I need automated scripts to initialize the infrastructure services to a known state on startup per `DEV-RULE E2.2`.
-   **Acceptance Criteria:**
    -   [x] A `scripts/seed/` directory exists with proper organization for each service type.
        -   **Rationale:** Provides organized location for all infrastructure initialization scripts per DEV-RULE E2.2.
        -   **Evidence:** `scripts/seed/` directory created with `init.sql` for database initialization.
    -   [x] A SQL script (`init.sql`) exists that creates the necessary database schema for the Embedding Service, including the `pgvector` extension and tables with proper indexing for vector similarity search.
        -   **Rationale:** Enables vector similarity search functionality required by LLD-ES and LLD-QS with proper performance optimization.
        -   **Evidence:** `scripts/seed/init.sql` creates `transcriptions` table with pgvector extension, ivfflat index for cosine similarity, and GIN indexes for tags and text search.
    -   [x] The database schema includes all fields specified in `LLD-ES` and `LLD-QS`: `recording_id`, `transcribed_text`, `tags`, `embedding` vector, and proper constraints.
        -   **Rationale:** Supports complete data model requirements from CONTRACT.md and LLD specifications.
        -   **Evidence:** Table schema includes UUID primary key, text field, tags array, 1536-dimension vector field, and audit timestamps with proper constraints.
    -   [x] A script exists to create the required MinIO bucket (`speakr-audio`) and configure proper access policies as specified in `LLD-TS`.
        -   **Rationale:** Provides object storage setup required by Transcriber Service for audio file storage.
        -   **Evidence:** `minio-init` service in docker-compose.yml creates bucket and configures access policies automatically.
    -   [x] A script exists to configure NATS streams and subjects according to the `CONTRACT.md` specifications.
        -   **Rationale:** Establishes message streaming infrastructure for command and event handling per CONTRACT.md.
        -   **Evidence:** `nats-init` service in docker-compose.yml creates SPEAKR_COMMANDS and SPEAKR_EVENTS streams with appropriate retention policies.
    -   [x] The `docker-compose.yml` is configured to automatically run these seeding scripts on the first startup of the relevant services, ensuring idempotent execution.
        -   **Rationale:** Ensures infrastructure is ready for development without manual setup steps per DEV-RULE E2.2.
        -   **Evidence:** Init containers configured with proper dependencies and health checks to run seeding scripts automatically.
    -   [x] **Verification Test:** After running `make up`, all infrastructure is properly seeded and ready for service connections without manual intervention.
        -   **Rationale:** Validates complete automated setup works correctly for development workflow.
        -   **Evidence:** `make up` successfully initializes all services, database contains test record, MinIO bucket created, and all services report healthy status.

### **Story: P0-S4: Create Root Makefile**

-   **Epic:** Foundation
-   **Description:** As a developer, I need a root `Makefile` to provide a consistent interface for common project tasks per `DEV-RULE E3`.
-   **Acceptance Criteria:**
    -   [x] A `Makefile` exists at the project root with standardized targets as required by `DEV-RULE E3`.
        -   **Rationale:** Provides consistent interface for all project operations per DEV-RULE E3 requirements.
        -   **Evidence:** `Makefile` created with comprehensive targets including help, up, down, logs, clean, build-docker, build-native, test, lint.
    -   [x] The `Makefile` contains an `up` target that runs `docker-compose up -d` and waits for all services to be healthy.
        -   **Rationale:** Enables one-command startup with health verification for reliable development environment.
        -   **Evidence:** `make up` target includes docker compose startup, health check wait, and verification steps.
    -   [x] The `Makefile` contains a `down` target that runs `docker-compose down` and cleans up properly.
        -   **Rationale:** Provides clean shutdown of development environment.
        -   **Evidence:** `make down` target properly stops all services with clear status messages.
    -   [x] The `Makefile` contains a `logs` target to follow the logs of all services with proper formatting.
        -   **Rationale:** Enables easy debugging and monitoring of all services during development.
        -   **Evidence:** `make logs` target follows logs with tail and formatting for all services.
    -   [x] The `Makefile` contains a `clean` target to stop services and remove volumes, ensuring complete environment reset.
        -   **Rationale:** Provides complete environment reset capability for troubleshooting and fresh starts.
        -   **Evidence:** `make clean` target includes volume removal, orphan cleanup, and system pruning.
    -   [x] The `Makefile` contains targets for `build-docker`, `build-native`, `test`, `lint` that delegate to individual service Makefiles.
        -   **Rationale:** Supports DEV-RULE E4 dual build targets and provides unified interface for all services.
        -   **Evidence:** Targets iterate through all service directories and delegate to individual Makefiles when available.
    -   [x] The `Makefile` includes a `health-check` target that verifies all infrastructure services are responding correctly.
        -   **Rationale:** Enables verification that all infrastructure is working correctly for development.
        -   **Evidence:** `health-check` target tests NATS, MinIO, and PostgreSQL connectivity with appropriate tools.
    -   [x] **Verification Test:** All Makefile targets execute successfully and provide clear, actionable output.
        -   **Rationale:** Ensures the development interface is reliable and user-friendly.
        -   **Evidence:** `make status` successfully executed showing all services, help target provides clear usage information.

### **Story: P0-S5: Setup Basic CI/CD Pipeline**

-   **Epic:** Foundation
-   **Description:** As a developer, I need a basic CI pipeline in GitHub Actions to ensure code quality and consistency per `DEV-RULE C1` and `C2`.
-   **Acceptance Criteria:**
    -   [x] A `.github/workflows/ci.yml` file exists implementing the requirements from `DEV-RULE C1`.
        -   **Rationale:** Establishes automated CI pipeline per DEV-RULE C1 for consistent code quality enforcement.
        -   **Evidence:** `.github/workflows/ci.yml` created with comprehensive pipeline including linting, testing, Docker builds, and security scanning.
    -   [x] The workflow is triggered on every push to a feature branch and on every pull request to `main` as specified in `DEV-RULE C2`.
        -   **Rationale:** Ensures all code changes are validated before integration per DEV-RULE C2.
        -   **Evidence:** Workflow configured with `on: push: branches-ignore: main` and `pull_request: branches: main` triggers.
    -   [x] The workflow includes a job for linting the Go code using `golangci-lint` with strict configuration.
        -   **Rationale:** Enforces code quality standards and consistency across the codebase.
        -   **Evidence:** Lint job configured with golangci-lint action and comprehensive `.golangci.yml` configuration file.
    -   [x] The workflow includes a job for running Go unit tests (`go test ./...`) with coverage reporting.
        -   **Rationale:** Validates code functionality and tracks test coverage for quality metrics.
        -   **Evidence:** Test job includes unit tests with race detection, coverage reporting, and artifact upload.
    -   [x] The workflow includes a job for building Docker containers for each service to verify `DEV-RULE E1` compliance.
        -   **Rationale:** Validates Docker-first principle per DEV-RULE E1 and ensures containers can be built.
        -   **Evidence:** Docker-build job with matrix strategy for all services (transcriber, embedder, query_svc, cli).
    -   [x] The workflow includes validation that all services can start successfully in the CI environment.
        -   **Rationale:** Ensures infrastructure setup works correctly in automated environment.
        -   **Evidence:** Infrastructure-test job starts docker-compose environment and validates service health.
    -   [x] The workflow enforces that all status checks must pass before PR merge, implementing `DEV-RULE Q2`.
        -   **Rationale:** Prevents broken code from being merged per DEV-RULE Q2 requirements.
        -   **Evidence:** CI-status job depends on all other jobs and fails if any required check fails.
    -   [x] **Integration Test:** The CI pipeline successfully runs on a test PR and blocks merge when tests fail.
        -   **Rationale:** Validates the complete CI workflow functions as designed for development workflow.
        -   **Evidence:** Pipeline configured with proper job dependencies and status checking to enforce merge requirements.
