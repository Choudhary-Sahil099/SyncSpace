# SyncSpace

> **A real-time collaborative workspace for multiple users to edit and interact with the same document simultaneously.**

SyncSpace is a real-time collaborative platform designed to demonstrate how modern collaborative editors handle  **concurrent users, synchronization, presence, cursor tracking, document state, and conflicting edits** .

The project uses a **Go-based concurrent WebSocket server** as the collaboration engine and a **React + TypeScript client** for the user interface.

The primary goal of SyncSpace is to understand and implement the core infrastructure behind systems such as Google Docs, Notion, and collaborative IDEs.

---

## ✨ Features

### 🔄 Real-Time Collaboration

Multiple users can connect to the same room and collaborate on a shared document in real time.

* WebSocket-based communication
* Persistent connections between clients and server
* Real-time document synchronization
* Room-based collaboration
* Broadcast updates to connected users
* Concurrent client handling

---

### 👥 Presence System

SyncSpace keeps track of users currently connected to a collaboration room.

The server maintains information about:

* Connected users
* User identity
* Room membership
* Connection/disconnection events
* User activity

This provides the foundation for a live collaboration experience.

---

### 🖱️ Live Cursor & Selection Tracking

Users can see the activity of other collaborators while editing.

The collaboration system tracks:

* Cursor position
* Text selection
* User presence
* Cursor updates
* Selection changes

This allows the frontend to represent where other users are currently working.

---

### 🏠 Collaboration Rooms

Users collaborate through isolated rooms.

Each room contains its own:

```text
Room
 ├── Connected Users
 ├── Document
 ├── Document Version
 ├── Operations
 └── Presence State
```

Users connected to one room receive updates from that room without affecting users in other rooms.

Example:

```text
Room: project-alpha

User A ─────┐
User B ─────┼──> Shared Document
User C ─────┘
```

---

## ⚡ Concurrent WebSocket Server

The backend is implemented in Go and designed around Go's concurrency primitives.

The server handles multiple WebSocket connections concurrently.

Conceptually:

```text
                 ┌───────────────┐
                 │   Go Server   │
                 └───────┬───────┘
                         │
             ┌───────────┼───────────┐
             │           │           │
             ▼           ▼           ▼
          Client A    Client B    Client C
             │           │           │
             └───────────┼───────────┘
                         │
                    Shared Room
                         │
                  Shared Document
```

The server is responsible for coordinating communication between clients rather than allowing clients to communicate directly with each other.

---

# 🧠 Conflict Handling

One of the main challenges in collaborative editing is handling concurrent modifications.

For example, suppose two users start with:

```text
Hello World
```

User A inserts:

```text
Beautiful 
```

while User B simultaneously deletes part of:

```text
World
```

If the server simply applies operations in the order they arrive, the final document can become inconsistent across clients.

SyncSpace therefore includes the foundation for  **Operational Transformation (OT)** .

---

## 🔀 Operational Transformation

The project implements OT transformation logic for the core operation combinations:

```text
Insert vs Insert
Insert vs Delete
Delete vs Insert
Delete vs Delete
```

Operations can be transformed against concurrent operations so that clients can converge toward the same document state.

Conceptually:

```text
Client A
   │
   │ Operation A
   ▼
Server ────────> Transform
   ▲
   │ Operation B
   │
Client B
```

The goal is:

```text
Client A Final State
          │
          ▼
       Same State
          ▲
          │
Client B Final State
```

This is one of the fundamental concepts behind real-time collaborative editors.

---

# 📄 Document Versioning

The server maintains document versions to help identify whether a client is working with an outdated document state.

A simplified version flow looks like:

```text
Version 10
    │
    ├── Client A operation
    │
    ▼
Version 11
    │
    ├── Client B operation
    │
    ▼
Version 12
```

When an operation arrives, its associated document version can be compared against the server's current version.

This provides the foundation for:

* Conflict detection
* Operation transformation
* Synchronization
* Recovery from stale client state

---

# 🛡️ Conflict Detection & Recovery

The server detects situations where a client submits an operation based on an older document version.

Example:

```text
Server Version = 10

Client A → operation based on Version 10
Client B → operation based on Version 10

Client A arrives first
        ↓
Server → Version 11

Client B arrives
        ↓
Client B is still based on Version 10
        ↓
Conflict detected
        ↓
Transform / reconcile operation
```

This is an important component of maintaining consistency in a distributed collaborative system.

---

# 💓 WebSocket Heartbeats

The server maintains long-lived WebSocket connections using heartbeat/ping mechanisms.

This helps detect:

* Disconnected clients
* Broken connections
* Stale connections
* Network failures

The system can therefore clean up users that are no longer connected to the collaboration room.

---

# 🔌 WebSocket Communication

The client establishes a WebSocket connection with the server.

A simplified connection looks like:

```text
Client
   │
   │ WebSocket connection
   ▼
Go WebSocket Server
   │
   │
   ├── Authentication
   ├── Room lookup
   ├── Connection registration
   ├── Presence updates
   ├── Document updates
   └── Broadcast
```

Example endpoint:

```text
/ws/<room>
```

The WebSocket layer is responsible for real-time communication while Gin handles the HTTP routing layer.

---

# 🏗️ Architecture

```text
                         SyncSpace
                            │
              ┌─────────────┴─────────────┐
              │                           │
          Frontend                    Backend
              │                           │
      React + TypeScript              Go + Gin
              │                           │
              │                       WebSocket
              │                           │
              └──────────────┬────────────┘
                             │
                       Collaboration
                           Engine
                             │
                ┌────────────┼────────────┐
                │            │            │
             Rooms       Presence     Documents
                │            │            │
                └────────────┼────────────┘
                             │
                         OT Engine
                             │
                    Conflict Resolution
```

---

# 🛠️ Technology Stack

## Frontend

* React
* TypeScript
* HTML5
* CSS3
* WebSocket API

## Backend

* Go
* Gin
* Gorilla WebSocket
* Go concurrency primitives
* WebSocket protocol

## Core Concepts

* Real-time distributed systems
* WebSockets
* Concurrent programming
* Operational Transformation
* Document versioning
* Conflict resolution
* Presence systems
* Event-based communication

---

# 📁 Project Structure

A simplified view of the project:

```text
SyncSpace/
│
├── client/
│   ├── src/
│   │   ├── components/
│   │   ├── hooks/
│   │   ├── services/
│   │   └── ...
│   │
│   ├── package.json
│   └── ...
│
├── server/
│   ├── ...
│   ├── go.mod
│   └── ...
│
├── docker-compose.yml
│
└── README.md
```

The exact internal structure may evolve as additional collaboration functionality is implemented.

---

# 🚀 Running the Project

## Prerequisites

Install:

* Go
* Node.js
* npm
* Docker / Docker Compose (if using the containerized setup)

---

## 1. Clone the Repository

```bash
git clone <repository-url>

cd SyncSpace
```

---

## 2. Start the Backend

Navigate to the server:

```bash
cd server
```

Install Go dependencies:

```bash
go mod download
```

Run the server:

```bash
go run .
```

The Go server will start and expose its HTTP/WebSocket endpoints.

---

## 3. Start the Frontend

Open another terminal:

```bash
cd client
```

Install dependencies:

```bash
npm install
```

Start the development server:

```bash
npm run dev
```

Open the frontend in your browser.

---

# 🧪 Testing

The backend contains unit tests for the collaboration and transformation logic.

Run:

```bash
go test ./...
```

For verbose output:

```bash
go test -v ./...
```

The OT transformation tests cover the primary combinations:

```text
Insert × Insert
Insert × Delete
Delete × Insert
Delete × Delete
```

These tests are important because incorrect transformation logic can cause different clients to end up with different document states.

---

# 🔬 Example Collaboration Flow

Consider three users joining:

```text
             ┌──────────────┐
             │    Room 1    │
             └──────┬───────┘
                    │
        ┌───────────┼───────────┐
        │           │           │
        ▼           ▼           ▼
     User A       User B       User C
        │           │           │
        └───────────┼───────────┘
                    │
             Shared Document
```

User A makes an edit:

```text
Hello World
     ↓
Hello SyncSpace World
```

The operation is sent to the server.

```text
User A
  │
  │ Operation
  ▼
Server
  │
  ├── Validate
  ├── Check version
  ├── Transform if necessary
  ├── Update document
  └── Broadcast
       │
       ├──────────> User B
       └──────────> User C
```

Users B and C then receive the resulting update.

---

# 🎯 Design Goals

SyncSpace is designed around several distributed-system principles:

### 1. Server-authoritative state

The server maintains the authoritative shared document state.

### 2. Real-time communication

Changes should propagate to collaborators with minimal latency.

### 3. Concurrency

Multiple users must be able to perform operations simultaneously.

### 4. Consistency

Collaborators should converge toward the same document state.

### 5. Fault tolerance

Disconnected or stale clients should not permanently corrupt the shared state.

### 6. Extensibility

The architecture should allow additional collaboration and AI capabilities to be added later.

---

# 📊 Current Implementation

| Capability                        | Status     |
| --------------------------------- | ---------- |
| React client                      | ✅         |
| TypeScript client                 | ✅         |
| Go backend                        | ✅         |
| Gin HTTP server                   | ✅         |
| WebSocket communication           | ✅         |
| Collaboration rooms               | ✅         |
| Multiple concurrent clients       | ✅         |
| Presence tracking                 | ✅         |
| Cursor tracking                   | ✅         |
| Selection tracking                | ✅         |
| Heartbeat / connection management | ✅         |
| Document state                    | ✅         |
| Document versioning               | ✅         |
| Conflict detection                | ✅         |
| Conflict recovery foundation      | ✅         |
| OT transformation engine          | ✅         |
| Insert/Insert transformation      | ✅         |
| Insert/Delete transformation      | ✅         |
| Delete/Insert transformation      | ✅         |
| Delete/Delete transformation      | ✅         |
| Backend unit tests                | ✅         |
| Operation-based frontend editing  | 🚧         |
| Production authentication         | 🚧         |
| PostgreSQL persistence            | 📋 Planned |
| Redis Pub/Sub                     | 📋 Planned |
| Production deployment             | 📋 Planned |
| E2E testing                       | 📋 Planned |

---

# 🗺️ Roadmap

## Phase 1 — Collaboration Core

* [X] WebSocket server
* [X] Room management
* [X] Presence
* [X] Cursor tracking
* [X] Selection tracking
* [X] Heartbeats
* [X] Document synchronization
* [X] Version tracking
* [X] Conflict detection
* [X] OT transformation engine
* [X] Unit testing

## Phase 2 — Production Infrastructure

* [ ] Operation-based frontend editing
* [ ] PostgreSQL document persistence
* [ ] Redis Pub/Sub for distributed WebSocket servers
* [ ] Production-grade JWT authentication
* [ ] Structured logging
* [ ] Graceful server shutdown
* [ ] Integration tests
* [ ] End-to-end tests
* [ ] Dockerized production deployment

## Phase 3 — Advanced Collaboration

* [ ] Rich text editor
* [ ] Undo / redo
* [ ] Document history
* [ ] Version timeline
* [ ] Better conflict visualization
* [ ] Offline editing
* [ ] Reconnection and state recovery

## Phase 4 — AI Collaboration

The long-term direction is to turn SyncSpace into an  **AI-powered collaborative workspace** .

Planned capabilities include:

* AI collaboration assistant
* Semantic conflict resolution
* Smart writing suggestions
* Live AI document review
* AI meeting summaries
* Voice collaboration
* Document knowledge graph
* Semantic document search
* Timeline replay
* Document memory
* AI code mode
* Plugin / extension ecosystem

---

# 🧩 Why SyncSpace?

Real-time collaboration looks simple from the user's perspective:

```text
Type → Other user sees it
```

But underneath, it requires solving difficult distributed-systems problems:

```text
Concurrent Users
       ↓
Network Communication
       ↓
Ordering
       ↓
Concurrent Operations
       ↓
Conflict Resolution
       ↓
State Synchronization
       ↓
Consistency
```

SyncSpace was built to explore these problems from the ground up rather than relying entirely on an existing collaboration service.

---

# 📚 Engineering Concepts Demonstrated

This project provides practical experience with:

* Distributed systems
* Concurrent programming
* WebSockets
* Event-driven architecture
* Client-server synchronization
* Operational Transformation
* Conflict resolution
* State management
* Connection lifecycle management
* Presence systems
* Versioned documents
* Automated testing
* Containerized development

---

# 🤝 Contributing

Contributions are welcome.

A typical workflow:

```bash
git checkout -b feature/my-feature

# Make changes

go test ./...

git add .
git commit -m "Add my feature"

git push origin feature/my-feature
```

Then open a pull request.

---

# 📜 License

This project is intended as an educational and engineering project for exploring real-time collaborative systems.

Add the project's chosen license here when the repository is finalized.

---

## ⭐ Project Summary

**SyncSpace is a real-time collaborative platform built with React, TypeScript, Go, Gin, and WebSockets. It implements room-based collaboration, presence and cursor tracking, document versioning, concurrent editing, conflict detection, and an Operational Transformation engine for reconciling concurrent edits.**

The project focuses on understanding the distributed-systems challenges behind real-time collaborative applications and provides a foundation for future AI-powered collaboration capabilities.
