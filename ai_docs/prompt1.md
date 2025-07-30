# Project: Modular, Channel-Agnostic Bot Engine

## 1. Description
This document specifies the requirements for a new, channel-agnostic bot engine written in Go. The architecture will be designed from the ground up to support multiple input and output channels (e.g., Slack, REST API, MCP integration, alerts) via a pluggable system. All message and event processing will flow through a standardized middleware pipeline, ensuring consistent logic application across all communication channels.

## Goal statement
To build a flexible and extensible bot engine where the core logic is completely decoupled from the communication channels, allowing for the seamless integration of new input and output sources.

## Core Architectural Concepts
The system will be built around three core concepts:
- **Bot Engine (The 'NPC' Core):** This central component is responsible for orchestrating all logic. It will manage actions (the bot's capabilities), and the middleware pipeline. It will be completely agnostic of any communication protocol.
- **Middleware Pipeline:** All incoming requests and outgoing responses will pass through a chain of middleware components. This allows for cross-cutting concerns like logging, authentication, authorization, and request modification to be handled in a clean, reusable way.
- **Communication Channels:** These are pluggable modules that handle the specifics of interacting with external services like Slack, a REST API, etc. Each channel will implement a standardized `Communication` interface, allowing the NPC core to interact with them uniformly.

## context & problem definition

### problem statement
Building a bot that is tightly coupled to a single communication platform (like Slack or Discord) results in a monolithic architecture. This makes it difficult to extend the bot's functionality to other platforms, reuse the core logic in different contexts, or integrate with other systems like internal APIs or alerting platforms. This project aims to solve this by creating a foundation that is modular and extensible by design.

### success criteria
- The initial build supports at least two different communication channels: Slack and a REST API.
- The core bot engine (`npc` package) has zero direct dependencies on any specific communication channel implementation.
- The middleware pipeline processes requests and responses from all integrated communication channels.
- The architecture is clearly documented and designed to be easily extensible, allowing a developer to add a new communication channel (e.g., Discord, MS Teams) with minimal effort.

## technical requirements

### functional requirements
- The system must be able to receive requests from multiple, concurrent sources (e.g., Slack, API).
- All requests must be processed through a configurable middleware pipeline.
- The system must be able to send responses back to the original source of the request.
- The core bot logic must be completely decoupled from the communication channels.

### non-functional requirements
- The new architecture should be performant and scalable.
- The code must be well-documented to facilitate future development.

### Technical Constraints
- The project must be written in Go.
- A middleware and action-based structure for processing logic must be implemented.

## Data & Database changes

### Data model updates
N/A

### Data migration plan
N/A

## API & Backend changes

### Data access pattern
The communication channels will be responsible for receiving requests and forwarding them to the `npc` core. The `npc` core will then process the requests and return a response to the communication channel.

### server actions
- A new `Communication` interface will be created to abstract the communication layer.
- Each communication channel (e.g., Slack, API) will implement the `Communication` interface.
- The `Npc` struct will be built to work with the `Communication` interface.

### API Routes
A new set of API endpoints will be created to allow for interaction with the bot via HTTP requests. For example:
- `POST /api/request`: This endpoint will accept a JSON payload with the request details and will return a JSON response.

## frontend changes

### New components
N/A

### Page updates
N/A

## Implementation plan

### Phase 1 - Core Engine and Interfaces
- **Goal:** Establish the foundational packages and interfaces.
- **Tasks:**
  - Create the `npc` package for the core bot engine.
  - Define the `Middleware` interface within the `npc` package.
  - Define the `Action` struct and registration system within the `npc` package.
  - Define the `Communication` interface in a new `channels` or `pkg` directory. This interface will define the methods all communication modules must implement (e.g., `Start()`, `Stop()`, `SendMessage()`, `RegisterRequestHandler()`).
  - Implement the core `Npc` struct in the `npc` package to work with the `Communication` interface.

### Phase 2 - Slack Communication Channel
- **Goal:** Implement the first communication channel for Slack.
- **Tasks:**
  - Create a new `slack` package (e.g., `channels/slack`).
  - Implement the `Communication` interface in this package, handling connection to Slack's Socket Mode API.
  - Implement the logic to translate Slack events into a generic request format for the NPC core. This includes packaging the Slack User ID into the request context.
  - Implement the logic to translate generic responses from the NPC core into Slack messages.

### Phase 3 - API Communication Channel
- **Goal:** Implement a REST API as a second communication channel.
- **Tasks:**
  - Create a new `api` package (e.g., `channels/api`).
  - Implement the `Communication` interface in this package.
  - Set up an HTTP server that listens for requests on a configurable port.
  - Define a standard JSON request/response format.
  - Implement handlers to translate HTTP requests into generic requests for the NPC core and vice-versa. This includes packaging security tokens from headers into the request context.

### Phase 4 - Main Application and Testing
- **Goal:** Integrate the components and ensure they work together.
- **Tasks:**
  - Create a `main` package and `main.go` file.
  - In `main`, initialize the NPC core, the Slack channel, and the API channel.
  - Implement initial authentication middleware for both Slack (e.g., user whitelist) and the API (e.g., static token check).
  - Start all communication channels to run concurrently.
  - Write unit tests for the core engine and each communication channel.
  - Write integration tests to verify that requests from both Slack and the API are processed correctly through the middleware pipeline.

## 5. Testing Strategy
### Unit Tests
- Write unit tests for the `Npc` struct to ensure it correctly interacts with the `Communication` interface and middleware.
- Write unit tests for the Slack and API communication channel implementations.
### Integration Tests
- Write integration tests to verify that the Slack and API communication channels can successfully send and receive messages through the `npc` core.
### End-to-End (E2E) Tests
- Define and (optionally) automate E2E tests that simulate a user interacting with the bot through both the Slack and API channels.

## 6. Security Considerations
### Authentication & Authorization
- Authentication and Authorization will be handled by dedicated middleware components within the core engine's pipeline.
- Each communication channel is responsible for extracting relevant security metadata from incoming requests (e.g., a Slack User ID, an HTTP `Authorization` header) and attaching it to the generic request context that is passed to the bot engine.
- Middleware will then inspect this context to perform validation. For example:
  - A `SlackAuthMiddleware` could check if a User ID is present in a predefined whitelist.
  - An `APIAuthMiddleware` could validate an API key against a list of known keys.
- This design allows for flexible and powerful security models. For instance, the `APIAuthMiddleware` could be extended in the future to integrate with a 3rd party authentication service (like OAuth or an internal identity provider) without changing the core engine or the API communication channel itself.
### Data Validation & Sanitization
- All incoming requests from the API channel should be validated and sanitized to prevent security vulnerabilities.
### Potential Vulnerabilities
- The new API endpoint could be a potential target for attacks. It is important to implement proper security measures to protect it.

## 7. Rollout & Deployment
### Deployment Steps
- The bot will be built and deployed as a single binary.
- Configuration for the bot engine and all communication channels (e.g., API keys, port numbers) will be managed via environment variables or a configuration file.
### Contingency Plan
- In case of critical bugs in production, the deployment will be rolled back. A hotfix will be prepared, tested, and deployed.

## 8. Open Questions & Assumptions
- What will be the initial list of authorized Slack users and/or API keys?
- What is the expected format of the API requests and responses?
