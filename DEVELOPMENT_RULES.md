# Speakr Development Rules

This document outlines the authoritative set of rules and processes for the development of the `speakr` project. Adherence to these rules is mandatory to ensure consistency, quality, and efficient parallel development.

---

### 1. Environment & Tooling

*   **E1: Docker-First Principle:** The official, production-ready artifact for any service **must** be a Docker image. The CI/CD pipeline will exclusively build and test using Docker.
*   **E2: Local Environment via Docker Compose:** The entire ecosystem (all services and their dependencies like NATS, MinIO, Postgres) **must** be orchestrated for local development using a single `docker-compose.yml` file at the project root.
    *   **E2.1: One-Command Startup:** A developer **must** be able to spin up the entire local environment with a single command (e.g., `make up` or `docker-compose up`).
    *   **E2.2: Automated Seeding:** The local environment startup **must** include scripts to automatically seed dependencies as needed (e.g., creating NATS streams, MinIO buckets, or database schemas).
*   **E3: Standardized `Makefile` Interface:** Every service **must** provide a `Makefile` with, at minimum, the following targets: `build-docker`, `build-native`, `test`, `lint`, `clean`.
*   **E4: Dual Build Targets:** To facilitate practical development and hardware interaction, the build system **must** support two build targets:
    *   **E4.1: `build-docker`:** This target **must** build the service within a Docker container, producing a production-ready Docker image.
    *   **E4.2: `build-native`:** This target **must** compile a native binary that runs directly on the host OS. By default, this binary **must** be configured to connect to the services running in the local Docker Compose environment (e.g., `NATS_URL=nats://localhost:4222`).

### 2. Development Workflow & Story Completion

*   **W1: Authoritative Development Plans:** All development work **must** be guided by a `PHASE_X_PLAN.md` document, which breaks down Epics into Stories with clear Acceptance Criteria (AC).
*   **W2: Definition of Done & Commit Process:** A commit for a completed story **must** only be made after the following process is completed, in order:
    1.  **Complete Development:** The code for the story is finished.
    2.  **Update the Plan Document:** The developer **must** edit the relevant `PHASE_X_PLAN.md` file.
    3.  **Mark Acceptance Criteria:** For every AC in the story, the developer **must** mark the checkbox as complete (`[x]`).
    4.  **Provide Evidence:** For every AC, the developer **must** write a brief `Rationale` explaining how the work meets the criteria and provide `Evidence` of its completion (e.g., a link to a passing CI build, a screenshot of log output, a sample cURL command and its response).
    5.  **Make the Commit:** Only after the plan document is updated and saved can the developer create the commit.

### 3. Version Control & Branching

*   **V1: Git Branching Model:** The project will use a simplified Gitflow model (`main`, `feature/*`, `bugfix/*`).
*   **V2: Conventional Commits & Story Association:**
    *   **V2.1:** All commit messages **must** adhere to the Conventional Commits specification.
    *   **V2.2:** The commit message for a completed story **must** be structured as follows: `feat(<service-name>): Complete story <Story-ID> - <Story Title>`. This creates a direct link between the code and the plan.

### 4. Code Quality & Pull Requests (PRs)

*   **Q1: All Changes via Pull Requests:** All changes to `main` **must** be made through a PR.
*   **Q2: Mandatory Status Checks:** A PR cannot be merged unless all automated status checks (linting, tests, container build) have passed.
*   **Q3: Mandatory Peer Review:** Every PR **must** be reviewed and approved by at least one other team member.

### 5. Testing Strategy

*   **T1: Unit Tests:** Core business logic **must** be covered by unit tests using mocks.
*   **T2: Integration Tests:** Each service **must** have tests that validate its interaction with its real dependencies.
*   **T3: Contract Tests:** The project **must** implement contract testing to ensure services adhere to `CONTRACT.md`.

### 6. CI/CD

*   **C1: CI/CD via GitHub Actions:** The project's CI/CD pipeline **must** be implemented using GitHub Actions.
*   **C2: Automated CI Pipeline:** The pipeline **must** automatically run all checks on every PR.
*   **C3: Automated Deployment to Staging:** Merges to `main` **must** trigger an automatic deployment to a `staging` environment.

### 7. Configuration & Secrets Management

*   **S1: Configuration via Environment Variables:** All service configuration **must** be supplied via environment variables.
*   **S2: Secure Secrets Management:** Secrets **must** be managed through a secure vault. For local development, secrets can be loaded from a `.env` file, which **must** be in `.gitignore`.

### 8. Observability

*   **O1: Structured Logging:** All log output **must** be in JSON format.
*   **O2: Traceability in Logs:** Every log entry **must** include the `correlation_id` and `recording_id` where applicable.
*   **O3: Health Check Endpoints:** Every service **must** expose a `/health` endpoint.
