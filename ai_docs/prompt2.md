# Project: Strongly Typed Request Object

## 1. Description
This document outlines the requirement to refactor the bot engine to use a strongly typed `Request` object instead of a generic `interface{}` for incoming requests. This change will improve type safety, code readability, and maintainability, making it easier to integrate new communication channels without modifying core logic.

## Goal statement
To enhance the modularity and extensibility of the bot engine by introducing a well-defined `Request` struct that standardizes the format of incoming data from all communication channels.

## Core Architectural Concepts
- **Generic Request Struct**: A new `Request` struct will be defined in the `npc` package to hold all necessary information about an incoming request, including the intended action, sender details, and raw payload.
- **Channel-Specific Adapters**: Each communication channel will be responsible for adapting its native incoming data format into an instance of the generic `Request` struct.
- **Unified Processing**: The `Middleware` pipeline and `Action` handlers will operate on this standardized `Request` struct, ensuring consistent processing across all channels.

## Context & Problem Definition

### Problem Statement
The current middleware design, where `Middleware.Execute` returns a `Response` object containing a potentially modified `Request` in its `Data` field, is not idiomatic Go. It complicates in-place modifications of the request object as it flows through the chain and makes error handling less direct. This approach also makes it harder to reason about the state of the request at different points in the pipeline.

### Core Architectural Concepts
- **Generic Request Struct**: A new `Request` struct will be defined in the `npc` package to hold all necessary information about an incoming request, including the intended action, sender details, raw payload, and a new `Args` field for arbitrary key-value arguments.
- **Centralized Middleware Chaining (Revised)**: The `npc` core will be responsible for managing and executing the middleware chain internally.
    - `Middleware.Execute` will now accept a pointer to the `Request` (`*Request`) and return only an `error`. This allows middleware to modify the request in place.
    - If a middleware returns an `error`, the chain will stop immediately.
    - If no error is returned, the chain continues with the potentially modified `*Request`.
- **Action Handlers**: Action handlers will continue to receive a `Request` (not a pointer) as they are terminal and do not modify the request for subsequent middleware.
- **Channel-Specific Adapters**: Each communication channel will be responsible for adapting its native incoming data format into an instance of the generic `Request` struct.
- **Unified Processing**: The `Middleware` pipeline and `Action` handlers will operate on this standardized `Request` struct, ensuring consistent processing across all channels.

### Success Criteria
- All `Middleware` handlers receive a pointer to a `Request` object and return an `error`.
- All `Action` handlers receive a `Request` object and return a `Response`.
- Each communication channel correctly transforms its incoming data into the `Request` format.
- The `npc` core handles the middleware chaining internally, simplifying middleware implementation.
- The `Request` struct includes a `Args` field (map[string]string) for key-value arguments.
- A new `AuditLogMiddleware` successfully logs request details.
- The core bot engine (`npc` package) remains unaware of the specific data formats of individual communication channels.
- The refactored code is clear, type-safe, and demonstrably easier to extend with new channels.

## Technical Requirements

### Functional Requirements
- The system must process incoming requests using the new `Request` struct.
- Each communication channel must provide a mechanism to convert its native request format into the `Request` struct.
- The `Request` struct must contain fields for `Action` (string), `User` (string, e.g., user ID), `ChannelID` (string), `Text` (string, textual representation of the incoming request payload), `Source` (string, e.g., "API", "Slack"), `AuthMethod` (string, e.g., "apikey", "slack_user"), `AuthToken` (string, the actual token or user ID), `Args` (map[string]string for arbitrary key-value arguments), and `RawData` (interface{} for the original payload).
- A new `AuditLogMiddleware` must be implemented to log the `Action`, `Source`, and `Text` of each request.

### Non-Functional Requirements
- The refactoring should not negatively impact performance.
- The code must remain well-documented.

### Technical Constraints
- The project must remain in Go.
- The existing action-based processing structure must be preserved.

## Data & Database Changes

### Data Model Updates
N/A

### Data Migration Plan
N/A

## API & Backend Changes

### Data Access Pattern
Communication channels will now construct a `Request` object from incoming data and pass it to the `npc` core. The `npc` core and its middleware/actions will operate solely on this `Request` object.

### Server Actions
- The `Communication` interface's `RegisterRequestHandler` will be updated to pass the `Request` object.
- The `Npc` struct's `ProcessRequest` method will accept a `Request` object and return a `Response`, internally managing the middleware chain by passing `*Request` to middleware.

### API Routes
No changes to API routes are expected, but the internal handling of the request payload will change to construct the `Request` object.

## Frontend Changes
N/A

## Implementation Plan

### Phase 1 - Define Request Struct and Update Core Interfaces
- **Goal**: Establish the new `Request` struct and update the `npc` package interfaces.
- **Tasks**:
  - Define the `Request` struct in `npc/request.go` (add `Args` field).
  - Update `npc/npc.go`:
    - Modify `Middleware` interface: `Execute(request *Request) error`.
    - Modify `Action` interface: `Handler(request Request) Response`.
    - Modify `ProcessRequest` method to accept `Request`, return `Response`, and implement centralized middleware chaining by iterating through middleware and passing `*Request`.

### Phase 2 - Update Middleware
- **Goal**: Adapt existing middleware and add new middleware.
- **Tasks**:
  - Modify `middleware/auth.go`: Update `Execute` to accept `*Request` and return `error`.
  - Create `middleware/audit_log.go`: Implement `AuditLogMiddleware` with `Execute(request *Request) error`.

### Phase 3 - Update Communication Channels
- **Goal**: Implement channel-specific transformation to the `Request` struct.
- **Tasks**:
  - Modify `channels/api/api.go`: In `handleRequest`, parse the incoming JSON into a `Request` object, including `Args`.
  - Modify `channels/slack/slack.go`: In `handleEvent`, parse the Slack event into a `Request` object, including `Args`.

### Phase 4 - Update Main Application and Tests
- **Goal**: Integrate the changes into the main application and update all tests.
- **Tasks**:
  - Modify `main.go` to create and pass `Request` objects, and to register the new `AuditLogMiddleware`.
  - Update `npc/npc_test.go`, `middleware/auth_test.go`, `channels/api/api_test.go`, `channels/slack/slack_test.go`, and `integration_tests/integration_test.go` to use the new `Request` struct and `npc.Response` (with `Data`, `Error`, and `Code` fields) consistently, and to reflect the new middleware chaining.

## 5. Testing Strategy
- **Unit Tests**: Update existing unit tests to reflect the new `Request` struct and middleware chaining, and ensure individual components function correctly.
- **Integration Tests**: Verify that requests from both Slack and API channels are correctly transformed into `Request` objects and processed through the new middleware pipeline and actions, including the `AuditLogMiddleware`.

## 6. Security Considerations
- The new `Request` struct will help standardize data validation and sanitization, as all incoming data will pass through a common structure.
- The `AuditLogMiddleware` will provide a clear record of requests for security auditing.

## 7. Rollout & Deployment
- No changes to rollout or deployment strategy are anticipated.

## 8. Open Questions & Assumptions
- What specific fields beyond `Action`, `User`, `ChannelID`, `Text`, `Source`, `AuthMethod`, `AuthToken`, `Args`, and `RawData` should be included in the generic `Request` struct for future extensibility?
