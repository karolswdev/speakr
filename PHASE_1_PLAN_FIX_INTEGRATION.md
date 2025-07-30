# Phase 1 Integration Fix Plan

**Version:** 1.0  
**Status:** Ready for Development  
**Created:** 2025-07-30  

---

## **Objective**

To fix the critical integration issues discovered during P1-INT1 testing and properly complete the end-to-end backend integration that was claimed to be working but is actually failing.

## **Current Status Assessment**

### **✅ What Actually Works**
- Infrastructure services (NATS, MinIO, PostgreSQL) are healthy and running
- All 3 backend services (transcriber, embedder, query_svc) build successfully
- Core service logic unit tests pass with mocks
- Basic service construction and configuration
- Docker Compose infrastructure setup

### **❌ Critical Issues Discovered**
- **Backend services are not running** - only infrastructure is running
- **Integration tests failing** - services not accessible for end-to-end testing
- **Query service 404 errors** - service not running on expected port 8080
- **No service orchestration** - no way to start all backend services together
- **Missing service definitions** - docker-compose.yml only has infrastructure

### **🔍 Evidence of Failures**
```
=== FAIL: TestP1_INT1_FullPipelineIntegration (33.03s)
    --- FAIL: FullPipelineTest - Timeout waiting for recording.started event
    --- FAIL: TagPropagationTest - Failed to decode query response: EOF  
    --- FAIL: PerformanceTest - All requests failed with status 404
```

---

## **Story: P1-INT1.1: Fix Backend Service Orchestration & Integration**

**Description:** As a developer, I need to fix the critical integration issues by properly orchestrating all backend services and ensuring they can communicate end-to-end to fulfill the complete contract specified in `CONTRACT.md`.

### **Root Cause Analysis**
The integration tests are failing because:
1. Backend services are implemented but not running
2. Docker Compose only starts infrastructure, not services
3. No service orchestration or startup coordination
4. Integration tests expect running services but find none

### **Acceptance Criteria**

#### **AC1: Service Orchestration**
- [ ] Add service definitions for transcriber, embedder, and query_svc to `docker-compose.yml`
  - **Rationale:** Enables all backend services to be started together with infrastructure for complete system operation
  - **Evidence:** `docker-compose.yml` contains service definitions with proper networking, dependencies, and health checks

#### **AC2: Service Startup & Health**  
- [ ] All backend services start successfully via `make up` and report healthy status
  - **Rationale:** Validates complete system startup and readiness for integration testing
  - **Evidence:** `make health-check` shows all services (NATS, MinIO, PostgreSQL, transcriber, embedder, query_svc) as healthy

#### **AC3: Service Communication**
- [ ] Services can communicate with each other through NATS and HTTP as designed
  - **Rationale:** Ensures the event-driven architecture and API endpoints work in containerized environment
  - **Evidence:** NATS message flow works between services and HTTP endpoints are accessible

#### **AC4: Environment Configuration**
- [ ] All services use proper environment variables and configuration in containerized setup
  - **Rationale:** Ensures services can connect to infrastructure and each other with correct settings
  - **Evidence:** Services start without configuration errors and connect to NATS, MinIO, PostgreSQL successfully

#### **AC5: End-to-End Pipeline Test**
- [ ] **Full Pipeline Test:** Successfully executes recording.start → transcription.run → embedding → query retrieval
  - **Rationale:** Validates the complete backend pipeline works end-to-end as specified in CONTRACT.md
  - **Evidence:** Integration test `testFullPipeline()` passes with all steps completing successfully

#### **AC6: Contract Compliance Test**
- [ ] **Contract Testing:** All CONTRACT.md commands and events work correctly with proper payload validation
  - **Rationale:** Ensures strict adherence to CONTRACT.md specifications for reliable service integration
  - **Evidence:** Integration test `testContractCompliance()` validates all commands with proper event responses

#### **AC7: Tag Propagation Test**
- [ ] **Tag Propagation:** Tags flow from initial commands through transcription, embedding, and are queryable via HTTP API
  - **Rationale:** Ensures data traceability and queryability throughout the processing pipeline
  - **Evidence:** Integration test `testTagPropagation()` demonstrates tags preserved and queryable end-to-end

#### **AC8: Metadata Handling Test**
- [ ] **Metadata Handling:** Metadata fields are preserved through event processing without core service modification
  - **Rationale:** Maintains data integrity and enables adapter-specific logic without interference
  - **Evidence:** Integration test `testMetadataHandling()` shows metadata unchanged through pipeline

#### **AC9: Error Propagation Test**
- [ ] **Error Propagation:** Invalid commands trigger appropriate error events without breaking the system
  - **Rationale:** Ensures robust operation with proper error handling and system resilience
  - **Evidence:** Integration test `testErrorPropagation()` demonstrates graceful error handling

#### **AC10: Performance Test**
- [ ] **Performance:** System handles concurrent requests with acceptable response times
  - **Rationale:** Validates system scalability and responsiveness under realistic load
  - **Evidence:** Integration test `testPerformance()` shows concurrent requests complete within time limits

#### **AC11: Complete Integration Validation**
- [ ] **Integration Test Suite:** All integration tests pass demonstrating complete backend platform functionality
  - **Rationale:** Provides comprehensive validation that the backend platform is production-ready
  - **Evidence:** `INTEGRATION_TEST=true go test -v -run TestP1_INT1` shows all tests passing

#### **AC12: Documentation & Evidence**
- [ ] **Workflow:** All work is committed with proper evidence and this document updated before commit
  - **Rationale:** Ensures development process compliance and accurate work documentation
  - **Evidence:** This document updated with actual test results and evidence before commit creation

---

## **Implementation Approach**

### **Phase 1: Service Orchestration**
1. Add service definitions to `docker-compose.yml`
2. Configure proper networking and dependencies
3. Add health checks for all services
4. Test service startup with `make up`

### **Phase 2: Integration Testing**
1. Start all services in containerized environment
2. Run integration tests and fix failures one by one
3. Validate each acceptance criteria with actual evidence
4. Document real test results (not claims)

### **Phase 3: Validation & Documentation**
1. Run complete test suite and capture results
2. Update this document with actual evidence
3. Commit only when all tests actually pass

---

## **Success Criteria**

**P1-INT1.1 is complete when:**
- All backend services start via `make up`
- All integration tests pass with real evidence
- Complete backend pipeline works end-to-end
- All CONTRACT.md functionality is validated
- System is ready for Phase 2 CLI development

**Evidence Required:**
- Test output showing all integration tests passing
- Service logs showing successful startup and communication
- HTTP responses showing query service working
- NATS message flow demonstrating event-driven architecture

---

## **Notes**

This story addresses the gap between what was claimed to be working and what actually works. The goal is to achieve genuine Phase 1 completion with verifiable evidence, not false claims.