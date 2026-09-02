# Go Task Manager API
#tests ci
A RESTful Task Management API built in **Go**, using **Gin**, **GORM**, and **PostgreSQL**, containerized with **Docker Compose**, and load-balanced across multiple instances with **Nginx**.

This project started as a Go fundamentals exercise and evolved step by step into a small, production-style backend — the goal was not just to build a CRUD API, but to understand *why* each architectural decision (layering, connection pooling, load balancing, migrations) is made in a real backend system.

---

## Table of Contents

- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [How It Works](#how-it-works)
- [Getting Started](#getting-started)
- [Environment Variables](#environment-variables)
- [API Endpoints](#api-endpoints)
- [Database Migrations](#database-migrations)
- [Design Decisions](#design-decisions)
- [Roadmap](#roadmap)

---

## Architecture

```text
                         Client
                           │
                           ▼
                         Nginx
                      Load Balancer
                           │
              ┌────────────┴────────────┐
              ▼                         ▼
           api-1                     api-2
              │                         │
              └────────────┬────────────┘
                           ▼
                         GORM
                           │
                           ▼
                      PostgreSQL
```

The API runs as two **stateless** instances (`api-1`, `api-2`) behind an Nginx load balancer. Neither instance is exposed publicly — Nginx is the single public entry point on port `80`. Both instances read and write to the same PostgreSQL database, so a task created through one instance is immediately visible through the other.

### Layered design

```text
Client → Nginx → Gin Router → Handler → Service → Repository → GORM → PostgreSQL
```

| Layer | Responsibility | Must NOT contain |
|---|---|---|
| **Handler** | HTTP concerns: binding JSON, status codes, response shaping | Business logic, GORM/SQL |
| **Service** | Business logic, use-case orchestration | Gin, HTTP status codes, GORM internals |
| **Repository** | Persistence via GORM, query logic | HTTP logic, business rules |

Keeping these boundaries strict means the business logic has no idea it's being served over HTTP, and the persistence layer has no idea what a "task" means to the business — it only knows how to store and retrieve one.

---

## Tech Stack

- **Go** 1.26+
- **Gin** — HTTP web framework
- **GORM** — ORM for PostgreSQL
- **PostgreSQL** 17
- **Docker** / **Docker Compose**
- **Nginx** — reverse proxy / load balancer

---

## Project Structure

```text
task-api/
│
├── db/
│   └── postgres.go        # GORM connection + pool configuration
│
├── task/
│   ├── model.go            # Task struct (GORM model)
│   ├── repository.go       # Persistence layer (GORM queries)
│   ├── service.go          # Business logic layer
│   └── handler.go          # HTTP handlers (Gin)
│
├── migrations/
│   ├── 0001_create_tasks_table.up.sql
│   └── 0001_create_tasks_table.down.sql
│
├── nginx/
│   └── nginx.conf          # Load balancer config (round robin)
│
├── docker-compose.yml
├── dockerfile
├── go.mod
├── go.sum
├── main.go                 # Composition root / dependency wiring
├── .env
└── .gitignore
```

---

## How It Works

1. A request hits **Nginx** on port `80`.
2. Nginx forwards it to either `api-1` or `api-2` using round-robin load balancing.
3. The **Handler** parses the HTTP request and calls the **Service**.
4. The **Service** applies any business rules and calls the **Repository**.
5. The **Repository** uses **GORM** to read/write the shared **PostgreSQL** database.
6. Because both instances share the same database, the response is consistent no matter which instance handled the request — this is what makes the API horizontally scalable and stateless.

### Task model

```go
type Task struct {
    ID          uint   `json:"id" gorm:"primaryKey"`
    Title       string `json:"title" gorm:"not null"`
    Description string `json:"description"`
    Completed   bool   `json:"completed" gorm:"not null;default:false"`
}
```

---

## Getting Started

### Prerequisites

- Docker & Docker Compose
- (Optional, for local dev without Docker) Go 1.26+

### Run with Docker Compose

```bash
git clone https://github.com/youssefahmed9000/go-task-manager-api.git
cd go-task-manager-api
docker compose up --build
```

Verify all services are healthy:

```bash
docker compose ps
```

Expected running services: `postgres`, `api-1`, `api-2`, `nginx` (and `migrate`, once migrations are wired in).

### Test it

```bash
# Health check (alternates between api-1 / api-2)
curl http://localhost/health

# Create a task
curl -X POST http://localhost/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "Learn GORM", "description": "Connect Go to PostgreSQL"}'

# List tasks
curl http://localhost/tasks
```

Because both instances share the same database, a task created via one request may be listed back via a different instance on the next request — this is deliberate, and is the proof that the load balancing setup works correctly.

---

## Environment Variables

Create a `.env` file in the project root:

```env
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=taskdb

DATABASE_URL=postgres://postgres:postgres@postgres:5432/taskdb
```

> **Note:** Inside Docker, services connect to Postgres via the service name `postgres:5432`. From your host machine, use `localhost:5432` instead.

`.env` is included in `.gitignore` and should never be committed with real credentials.

---

## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/health` | Returns service status and which instance handled the request |
| `GET` | `/` | Basic service info |
| `GET` | `/tasks` | List all tasks |
| `POST` | `/tasks` | Create a new task |

**Planned (see [Roadmap](#roadmap)):** `GET /tasks/:id`, `PUT /tasks/:id`, `PATCH /tasks/:id/complete`, `DELETE /tasks/:id`

### Example: Create a task

**Request**
```http
POST /tasks
Content-Type: application/json

{
  "title": "Learn GORM",
  "description": "Connect Go to PostgreSQL"
}
```

**Response**
```json
{
  "id": 1,
  "title": "Learn GORM",
  "description": "Connect Go to PostgreSQL",
  "completed": false
}
```

---

## Database Migrations

Schema changes are managed through versioned SQL migration files rather than GORM's `AutoMigrate()`.

**Why not `AutoMigrate()`?** Early in development, each API instance called `AutoMigrate()` on startup. With two instances starting concurrently, both tried to create the `tasks` table at the same time, producing a race condition:

```text
duplicate key value violates unique constraint "pg_class_relname_nsp_index"
```

`AutoMigrate()` was removed entirely. Instead, schema changes are applied **once**, before any API instance starts, via a dedicated migration step — keeping schema management separate from application startup.

```bash
migrate -path migrations -database "$DATABASE_URL" up
```

---

## Design Decisions

A few deliberate choices worth calling out:

- **No public ports on `api-1` / `api-2`** — only Nginx is exposed on the host. API containers communicate over an internal Docker network (`taskflow-network`).
- **Healthchecks everywhere** — both Postgres and the API instances have Docker healthchecks, and `depends_on` uses `condition: service_healthy` so nothing starts before its dependencies are actually ready (not just "container running").
- **Persistent Postgres volume** — data survives container restarts.
- **Multi-stage Dockerfile** — the final runtime image (`alpine:latest`) doesn't contain the Go compiler/toolchain, keeping the image small and reducing attack surface.
- **Connection pooling is explicit** — `MaxOpenConns=25`, `MaxIdleConns=10`, `ConnMaxLifetime=30m`, `ConnMaxIdleTime=5m` are configured deliberately rather than left at defaults.

---

## Roadmap

- [x] Go fundamentals → REST API with Gin
- [x] Layered architecture (Handler / Service / Repository)
- [x] PostgreSQL + GORM integration
- [x] Dockerized, multi-instance deployment behind Nginx
- [x] Healthchecks, persistent volume, internal network
- [ ] Dedicated one-shot migration mechanism (replacing `AutoMigrate`)
- [ ] Full CRUD (`GET/PUT/PATCH/DELETE /tasks/:id`)
- [ ] Repository interfaces, DTOs, centralized error handling, request validation
- [ ] Structured logging, graceful shutdown, request timeouts, rate limiting, CORS
- [ ] Unit and integration tests (service, repository, handler, Postgres)
- [ ] Load testing and performance tuning
- [ ] High availability: read replicas, caching (Redis), observability, Kubernetes

---

## License

Add a license of your choice (e.g. MIT) if you intend this to be publicly reusable.
