# Development Plan: Phase 3 - Production Deployment & Release

**Version:** 1.0
**Status:** Ready for Development

---

### **Objective**

To deploy the `speakr` ecosystem to a production environment, automate the release process, and publish the CLI application for end-users. This phase transitions the project from a development effort to a live, stable, and scalable service.

### **Referenced Documents**

-   **Vision & Requirements:** `VISION.md`
-   **Rules & Architecture:** `ARCHITECTURE.md`, `DEVELOPMENT_RULES.md`
-   **Contracts & Interfaces:** `CONTRACT.md`
-   **Low-Level Designs:** `LLD_TRANSCRIBER_SERVICE.md`, `LLD_EMBEDDING_SERVICE.md`, `LLD_QUERY_SERVICE.md`, `LLD_CLI_APPLICATION.md`

---

## **Epic: Infrastructure as Code (IaC)**

### **Story: P3-IAC1: Provision Production Infrastructure**

-   **Description:** As a DevOps engineer, I need to define the production environment (VPC, Kubernetes Cluster, etc.) using Terraform to ensure it is repeatable and auditable.
-   **Acceptance Criteria:**
    -   [ ] A `terraform/` directory is created at the project root.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] Terraform scripts are created to provision a Virtual Private Cloud (VPC) with public and private subnets.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] Terraform scripts are created to provision a managed Kubernetes cluster (e.g., EKS, GKE).
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] Terraform scripts are created to provision managed services for NATS, PostgreSQL (with pgvector), and Object Storage (S3).
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] Terraform scripts are created to provision a secure vault for secrets management (e.g., AWS Secrets Manager).
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The Terraform configuration is successfully applied to a cloud provider account without errors.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Workflow:** All work is committed following the process in `DEV-RULE W2`, including updating this document with rationale and evidence before the commit.
        -   **Rationale:**
        -   **Evidence:**

---

## **Epic: CI/CD for Release**

### **Story: P3-CD1: Automate Container Image Publishing**

-   **Description:** As a DevOps engineer, I need to extend the CI/CD pipeline to automatically build and publish Docker images for the backend services to a container registry.
-   **Acceptance Criteria:**
    -   [ ] A `.github/workflows/release.yml` file is created.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The workflow is triggered only on the creation of a Git tag matching the pattern `v*.*.*` (e.g., `v1.0.0`).
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The workflow includes a job that builds the Docker image for each backend service (`transcriber`, `embedder`, `query_svc`).
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The workflow pushes the built images to a container registry (e.g., Docker Hub, AWS ECR) and tags them with the Git tag version.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Workflow:** All work is committed following the process in `DEV-RULE W2`, including updating this document with rationale and evidence before the commit.
        -   **Rationale:**
        -   **Evidence:**

### **Story: P3-CD2: Automate CLI Binary Release**

-   **Description:** As a DevOps engineer, I need the release pipeline to automatically cross-compile the CLI application and attach the binaries to a GitHub Release.
-   **Acceptance Criteria:**
    -   [ ] The `release.yml` workflow includes a job for the CLI application.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The job uses a tool like `gox` to cross-compile the CLI for Linux, macOS, and Windows (amd64).
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The job creates a new GitHub Release for the Git tag.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The compiled binaries are successfully attached as artifacts to the GitHub Release.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Workflow:** All work is committed following the process in `DEV-RULE W2`, including updating this document with rationale and evidence before the commit.
        -   **Rationale:**
        -   **Evidence:**

### **Story: P3-CD3: Implement Automated Deployment**

-   **Description:** As a DevOps engineer, I need the release pipeline to automatically deploy the new container images to the production Kubernetes cluster.
-   **Acceptance Criteria:**
    -   [ ] Kubernetes manifest files (Deployments, Services, etc.) are created for each backend service.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The `release.yml` workflow includes a final job for deployment.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The job securely authenticates with the cloud provider and the Kubernetes cluster.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The job updates the Kubernetes deployments to use the newly published container image versions.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Workflow:** All work is committed following the process in `DEV-RULE W2`, including updating this document with rationale and evidence before the commit.
        -   **Rationale:**
        -   **Evidence:**

---

## **Epic: Release & Documentation**

### **Story: P3-DOC1: Write User Installation Guides**

-   **Description:** As a technical writer, I need to create clear, user-friendly installation guides for the `speakr` CLI on all supported platforms.
-   **Acceptance Criteria:**
    -   [ ] A `docs/` directory is created in the repository.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] A `docs/installation.md` file is created.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The document provides step-by-step instructions for downloading the binary from GitHub Releases and placing it in the system PATH for Linux, macOS, and Windows.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] The document clearly lists all prerequisites (`ffmpeg`, `xclip` on Linux).
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Workflow:** All work is committed following the process in `DEV-RULE W2`, including updating this document with rationale and evidence before the commit.
        -   **Rationale:**
        -   **Evidence:**

### **Story: P3-DOC2: Finalize Project Documentation**

-   **Description:** As a project lead, I need to finalize all project documentation for the v1.0 release.
-   **Acceptance Criteria:**
    -   [ ] The main `README.md` is updated with a link to the installation guide and a summary of the v1.0 features.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] A `CHANGELOG.md` file is created, summarizing the changes for the v1.0 release.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] All foundational documents (`ARCHITECTURE.md`, `CONTRACT.md`, etc.) are reviewed and marked as stable for v1.0.
        -   **Rationale:**
        -   **Evidence:**
    -   [ ] **Workflow:** All work is committed following the process in `DEV-RULE W2`, including updating this document with rationale and evidence before the commit.
        -   **Rationale:**
        -   **Evidence:**
