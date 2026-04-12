# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

TaskFlow is a full-stack task management app. Backend: Go REST API. Frontend: React TypeScript SPA. Database: PostgreSQL (via Docker).

## Development Commands

### Start full stack (Docker Compose)
```bash
make up          # docker compose up -d (production-oriented: APP_ENV=prd, see docker-compose.yml)
make down        # docker compose down
```
Compose runs **`application-prd.yml`**: only **`WEB_PORT`** (default 3000) is published; Postgres and the API stay on the internal network. Requires **`POSTGRES_PASSWORD`**, **`JWT_SECRET`**, and **`CORS_ALLOWED_ORIGINS`** in `.env`.

### Run backend
```bash
make backend     # Equivalent to: cd backend && go run ./cmd/server
```

Config is loaded from `backend/config/application-{APP_ENV}.yml` (default: `application-dev.yml`).
Secrets (`DATABASE_URL`, `JWT_SECRET`) stay in the repo root `.env` (or `backend/.env` to override) or the environment — never in the YAML files. Use the single `.env.example` at the repo root.
Set `APP_ENV=prd` to switch to `application-prd.yml`.

**Docker Compose:** web UI at **`http://localhost:3000`**. The API and Postgres are **not** exposed on the host; use **`/api`** via nginx on the same origin.

### Run frontend
```bash
make frontend    # Equivalent to: cd frontend && npm run dev
```
Vite dev server on port **5173**, proxies `/api/*` to the backend on **4000**.

### Backend tests
```bash
cd backend && go test ./...
```

### Frontend lint and build
```bash
cd frontend && npm run lint
cd frontend && npm run build
```

## Architecture

### Backend (`backend/`)

Strict layered architecture: **Handler → Service → Repository → SQL**, with a shared `model` package as the domain layer.

```
internal/model/      ← domain types (User, Project, Task, TaskPatch) — imported by all layers
internal/repository/ ← SQL via pgx; scans rows into model types
internal/service/    ← business logic + auth rules; repo interfaces in repos.go
internal/handler/    ← HTTP only; request parsing, DTO conversion, error mapping
```

- `cmd/server/main.go` — Composition root: wires config → db → repositories → services → handlers → server
- `internal/config/` — Env-var config (reads `.env` via godotenv). `CORS_ALLOWED_ORIGINS` is comma-separated and **must be set in production**.
- `internal/db/` — pgx pool setup + golang-migrate (migrations embedded via `embed.FS` in `migrations/embed.go`)
- `internal/middleware/` — JWT auth (`auth.go` injects `user_id` into context), request logging (`logging.go`)
- `internal/handler/dto.go` — All response shapes (DTOs). Conversion from model types happens here at the boundary.
- `internal/handler/validate.go` — Format/parse validation only (email, UUID). Enum/business-rule validation is the service's job.
- `internal/handler/errors.go` — Translates domain errors to HTTP status codes; logs `internal_error` with `request_id`.
- `internal/service/repos.go` — Repository interfaces (what service depends on, what repository must satisfy).
- `internal/auth/` — JWT sign/parse (`jwt.go`), bcrypt helpers (`password.go`, cost 12)
- `internal/server/routes.go` — All route definitions; all API routes under `/api` prefix; `/healthz` outside

**Layer contracts:**
- Repository → translates `pgx.ErrNoRows` and DB constraint errors to `errs.ErrNotFound` / `errs.ErrForbidden` / `errs.ErrEmailTaken` before returning. Service never imports pgx.
- Service → returns `model.*` types. Never imports encoding/json. `model.TaskPatch` carries typed fields parsed by the handler.
- Handler → never imports `repository`. Only imports `model` and `service` interfaces.

**Authorization model:** Project owner or task participant (creator/assignee) can access resources.

**CSV demo seed:** Runs on every API startup via `seed.ApplyIfNeeded`; inserts 10 users, 20 projects, 60 tasks **once per database** (tracked in `app_seed_state`). CSV files live only under **`backend/data/seed/`** (copied into the Docker image at `/app/data/seed/`). Override directory with **`SEED_CSV_DIR`** if needed.

### Frontend (`frontend/src/`)

- `App.tsx` — Root router; `RequireAuth` wraps all protected routes
- `contexts/AuthProvider.tsx` — Global auth state (user + JWT token); persisted to `localStorage`; 401 responses trigger logout
- `api/client.ts` — All REST calls; `ApiError` carries HTTP status + field-level validation errors
- `pages/` — `LoginPage`, `RegisterPage`, `ProjectsPage`, `ProjectDetailPage`
- `components/TaskSidePanel.tsx` — Large component (~28KB) handling task create/edit with all fields
- `components/ui/` — Base primitives (Button, Input, Label) built on Radix UI + CVA

### Database

Schema defined in `migrations/000001_init.up.sql`. Key tables: `users`, `projects`, `tasks` (with enums for status and priority). Indexes on foreign keys and status/priority columns.

## Key Design Decisions

- **`internal/model`** is the shared domain layer — repository scans into model types, service returns model types, handler converts model types to response DTOs. No layer imports another layer's package.
- Services depend on **interfaces** (defined in `service/repos.go`), not concrete repository types — enables testing with mocks.
- **`model.TaskPatch`** uses typed Go fields (`*uuid.UUID`, `*time.Time`, bool flags). JSON parsing and null-vs-absent disambiguation happens in the handler before the struct is passed to service.
- Frontend uses **React Context** (no Redux/Zustand) for auth state.
- Migrations are **embedded** in the binary (`embed.FS`) — no external SQL files needed at runtime.
- The Vite dev server **proxies** `/api/*` to the Go backend, so no CORS handling is needed during development.
