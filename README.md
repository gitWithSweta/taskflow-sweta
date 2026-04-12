# TaskFlow

Full documentation for reviewers and contributors: **overview**, **high-level design (HLD)**, **low-level design (LLD)**, **how to run**, **migrations**, **test login**, **API reference**, and **future work**.

**Table of contents**

1. [Overview](#1-overview)  
2. [High-level design (HLD)](#2-high-level-design-hld)  
3. [Low-level design (LLD)](#3-low-level-design-lld)  
4. [Architecture decisions and tradeoffs](#4-architecture-decisions-and-tradeoffs)  
5. [Running locally](#5-running-locally)  
6. [Running migrations](#6-running-migrations)  
7. [Test credentials](#7-test-credentials)  
8. [API reference](#8-api-reference)  
9. [What you would do with more time](#9-what-you-would-do-with-more-time)  
10. [Repository (GitHub) and secrets](#10-repository-github-and-secrets)

---

## 10. Repository (GitHub) and secrets

**Public repository name:** `taskflow-Sweta`

**Layout — single monorepo** (not split into two repositories). Everything ships in one tree:

```
taskflow-Sweta/
├── docker-compose.yml      # root Compose: postgres + API + web
├── Dockerfile              # multi-stage build for the Go API
├── .env.example            # committed template (no real secrets)
├── .gitignore              # excludes .env and node_modules / dist
├── backend/                # Go module (cmd/server, internal/, migrations/, data/seed/)
├── frontend/               # Vite + React SPA
└── README.md
```

If you ever split **backend** and **frontend** into separate GitHub repos, add a short “Related repo” link in each README; this submission uses the **monorepo** above.

**Secrets**

- **Commit** only `.env.example` (placeholders / safe local defaults for reviewers).
- **Never commit** `.env`, `backend/.env`, or any file containing real `JWT_SECRET`, production database passwords, or API keys.
- If `.env` or secrets were ever committed, **rotate** `JWT_SECRET`, database passwords, and any other leaked values **before** you submit or share the repo.

**Create the public GitHub repo and push** (run on your machine; replace `YOUR_GITHUB_USERNAME`):

```bash
cd /path/to/TaskFlow
git init
git add .
git status   # confirm .env and node_modules are NOT listed
git commit -m "Initial commit: TaskFlow monorepo"
git branch -M main
```

**Option A — GitHub CLI** ([`gh`](https://cli.github.com/) installed and logged in):

```bash
gh repo create taskflow-Sweta --public --source=. --remote=origin --push
```

**Option B — GitHub website**

1. New repository → name **`taskflow-Sweta`** → Public → **do not** add a README (you already have one).
2. Then:

```bash
git remote add origin https://github.com/YOUR_GITHUB_USERNAME/taskflow-Sweta.git
git push -u origin main
```

Clone URL for reviewers:

`git clone https://github.com/YOUR_GITHUB_USERNAME/taskflow-Sweta.git`

---

## 1. Overview

**What it is.** TaskFlow is a full-stack task manager. People register or sign in, own **projects**, add **tasks** to those projects, assign tasks to users, set priority and due dates, and move work through **todo → in_progress → done**.

**What it does in practice.** The browser loads a React SPA. All REST calls go to **`/api/...`** on the same host (port 3000 in Docker). The Go service validates JWTs, enforces who may see or change which project/task, persists to PostgreSQL, and returns JSON. Demo data can be loaded once from CSV so reviewers can log in without registering.

**Tech stack.**

| Area | Technology |
|------|------------|
| API | Go 1.22, Chi router, jackc/pgx, golang-migrate, golang-jwt, bcrypt (cost 12), slog (JSON to stdout) |
| Database | PostgreSQL 16 |
| Web UI | React 19, TypeScript, Vite, React Router 7, Tailwind CSS v4, Radix UI primitives |
| Local deployment | Docker Compose: `postgres`, `backend` (API), `web` (nginx: static files + reverse proxy) |

**Important URL shape.** Every JSON route is under **`/api`** (for example `POST /api/auth/login`). **`GET /healthz`** is on the API root (not under `/api`) and is also exposed through nginx at `http://localhost:3000/healthz`.

---

## 2. High-level design (HLD)

### 2.1 System context

- **Actors:** End user (browser), reviewer (curl/Postman), optional local developer running Go or Vite on the host.
- **External systems:** PostgreSQL (data store). No third-party auth or email in this version.

### 2.2 Logical containers (Compose)

| Container | Role |
|-----------|------|
| **postgres** | Holds all relational data. Not published on the host in the default Compose file; only reachable on the internal Docker network. |
| **backend** | Go HTTP server on port 4000 *inside* the network. Loads YAML from `/app/config`, secrets from env (`DATABASE_URL`, `JWT_SECRET`, `CORS_ALLOWED_ORIGINS`). Runs SQL migrations on startup, then CSV seed if needed, then serves Chi routes. |
| **web** | nginx serves the built SPA from `/usr/share/nginx/html` and proxies `/api` and `/healthz` to `backend:4000`. **Only this service publishes a host port** (default **3000**). |

So from the reviewer’s laptop: **one origin** (`http://localhost:3000`) for both UI and API; no CORS configuration is required for that path as long as `CORS_ALLOWED_ORIGINS` includes that origin (the sample `.env.example` does).

### 2.3 Request path (happy path)

1. Browser requests `https://localhost:3000/project` (or any SPA route). nginx serves `index.html` (SPA shell).
2. Browser calls `POST http://localhost:3000/api/auth/login` with JSON body. nginx forwards to the Go server with path `/api/auth/login` preserved.
3. API validates credentials, creates a **session row**, returns a **JWT** whose `jti` claim equals that session’s UUID.
4. Later requests send `Authorization: Bearer <jwt>`. Middleware verifies signature and expiry, parses `user_id`, then checks the session row still exists (so **logout** invalidates the token immediately even before expiry).

### 2.4 Major domain concepts

- **User** — identity, bcrypt password hash, unique email.
- **Project** — owned by one user; deleting a project deletes its tasks (CASCADE).
- **Task** — belongs to one project; has status, priority, optional assignee and due date, and a **creator** (used for delete permission together with project ownership).

### 2.5 Authorization (who can do what)

- **Project owner** — full CRUD on the project and its tasks (with some PATCH rules for sensitive fields).
- **Task assignee or creator** — can see the project and participate; can PATCH tasks they are allowed to touch per service rules; **delete task** is allowed for **project owner** or **task creator** only.

This is enforced in the **service** layer, not only in HTTP handlers.

---

## 3. Low-level design (LLD)

### 3.1 Backend layering

Strict order: **Handler → Service → Repository → SQL**. A shared **`internal/model`** package holds domain structs used everywhere.

| Layer | Responsibility | Typical packages |
|-------|----------------|------------------|
| **Handler** | HTTP only: read JSON/query, call service, map `model` → JSON DTOs, map errors → status codes. **Must not** import `repository`. | `internal/handler` |
| **Service** | Business rules, authorization, orchestration. Depends on **interfaces** defined in `internal/service/repos.go`, not concrete repos. **Must not** import `encoding/json` or pgx. | `internal/service` |
| **Repository** | Parameterized SQL via pgx; scan into `model` types; translate `pgx.ErrNoRows` and constraint violations into small domain errors (`errs` package). | `internal/repository` |
| **Model** | Plain structs and enums (e.g. `TaskPatch` for partial updates). | `internal/model` |

**Composition root:** `cmd/server/main.go` loads config, connects the DB, runs migrations, runs seed, constructs concrete repositories and services, passes them into handlers, starts Chi with graceful shutdown on SIGINT/SIGTERM.

### 3.2 Key backend packages (by folder)

- **`internal/config`** — Loads `application-{APP_ENV}.yml` (default `dev` on host, `prd` in Compose), overlays env vars (`DATABASE_URL`, `JWT_SECRET`, `HTTP_PORT`, `CORS_ALLOWED_ORIGINS`, `SEED_CSV_DIR`, etc.). Repo root `.env` is loaded first, then optional `backend/.env`.
- **`internal/db`** — Builds pgx pool from config; runs **embedded** migrations from `migrations/*.sql` via `migrations/embed.go`.
- **`internal/auth`** — `HashPassword` / `CheckPassword` (bcrypt cost 12); `SignToken` / `ParseToken` (JWT with `user_id`, `email`, `jti`, expiry aligned with session row).
- **`internal/middleware`** — `RequireAuth`: Bearer parse → JWT verify → session existence check → set `user_id` and session id in context. Structured request logging and trace IDs.
- **`internal/handler/errors.go`** — Maps `errs.ErrNotFound` → 404, `ErrForbidden` → 403, invalid credentials / unauthorized → 401, validation maps → 400 with `fields` object.
- **`internal/httpx`** — `WriteJSON`, `WriteValidation`, bounded `ReadJSON`.
- **`internal/server/routes.go`** — Registers `/healthz` and `/api/...` routes; wires CORS from config.

### 3.3 Database schema (summary)

Defined in versioned SQL under `backend/migrations/`:

- **`users`** — `id`, `name`, `email` (unique), `password_hash`, `created_at`.
- **`projects`** — `id`, `name`, `description`, `owner_id` → `users`, `created_at`.
- **`tasks`** — `id`, `title`, `description`, `status` (enum), `priority` (enum), `project_id`, `assignee_id` (nullable), `creator_id`, `due_date`, `created_at`, `updated_at`. Indexes on foreign keys used in filters.
- **`app_seed_state`** — Tracks whether static CSV demo seed has run (`key` = marker string).
- **`user_sessions`** — One row per issued JWT session; `id` matches JWT `jti`; `expires_at` matches token expiry; deleted on logout.

### 3.4 CSV seed (algorithm)

On every API startup, `seed.ApplyIfNeeded`:

1. If a row exists in `app_seed_state` for the fixed marker key, **return immediately** (idempotent across restarts).
2. Otherwise load `users.csv`, `projects.csv`, `tasks.csv` from, in order: **`SEED_CSV_DIR`** if set; else `data/seed/` relative to process cwd; else `data/seed/` next to the executable (Docker copies `backend/data/seed` to `/app/data/seed/`).
3. Insert users (hashing passwords), then projects (referencing owner emails from the CSV), then tasks (referencing project slot + owner email), then insert the marker row, all in **one transaction**.

If you change CSV after seed already ran, you must reset the DB or remove the marker row (see [Test credentials](#7-test-credentials)).

### 3.5 Frontend structure

- **`App.tsx`** — Routes; protected routes wrapped in a guard that requires auth context.
- **`contexts/AuthProvider.tsx`** — Holds current user + JWT; persists token (and user snapshot) in `localStorage`; global `fetch` wrapper clears session on `401` when a token was sent.
- **`api/client.ts`** — All REST paths are relative to **`/api`**. Central place for `ApiError` and field-level server validation.
- **`pages/`** — Login, register, project list, project detail (tasks).
- **`components/`** — Task side panel, modals, layout, markdown description editor, etc.; **`components/ui/`** — Button, Input, Label (Radix + CVA).

State for auth is **React Context** only (no Redux/Zustand).

### 3.6 Configuration and secrets

- **Non-secret tuning** — Pool sizes, timeouts, log level: YAML under `backend/config/application-*.yml`.
- **Secrets** — `JWT_SECRET`, database password, etc.: **environment only** (`.env` for local/Compose). YAML `jwt_secret` is empty in repo and must be overridden by env in production-style runs.

---

## 4. Architecture decisions and tradeoffs

- **Why layers?** Keeps SQL and HTTP concerns out of business rules; tests can mock repositories; handlers stay thin.
- **Why JWT + DB session?** JWT carries identity without a DB round-trip for parsing, but **session row** gives **revocable** logout and a single place to invalidate tokens.
- **Why `/api` prefix?** Avoids colliding with SPA routes like `/project` and `/login` when everything is served from one host. Tradeoff: spec examples that omit `/api` need a mental prefix when comparing.
- **Why embed migrations?** One binary + config dir is enough to run against an empty database; no separate migrate CLI in production path.
- **Why CSV seed?** Predictable demo for reviewers without a separate “bootstrap” script; marker table prevents duplicate users on restart.
- **What was left out or kept small:** No OpenAPI/Bruno artifact in-repo; no rate limiting; integration tests are optional and tag-gated; frontend bundle is not code-split by route.

---

## 5. Running locally

Assume **Docker Desktop** (or compatible engine) is installed and running. **Go and Node are not required** for the default path.

```bash
git clone https://github.com/YOUR_GITHUB_USERNAME/taskflow-Sweta.git
cd taskflow-Sweta
cp .env.example .env
docker compose up --build
```

- **`--build`** ensures images exist on first clone (images are not pushed to a registry in this workflow).
- Open **http://localhost:3000** in a browser.

Compose **requires** `POSTGRES_PASSWORD`, `JWT_SECRET`, and `CORS_ALLOWED_ORIGINS` in `.env` (the sample `.env.example` supplies working local values).

**Optional — run API on the host:** Install Go 1.22+, run PostgreSQL with a reachable `DATABASE_URL`, set `JWT_SECRET` in `backend/.env` (and load repo root `.env` if you use it), then:

```bash
cd backend && go run ./cmd/server
```

**Optional — Vite dev server:** `cd frontend && npm install && npm run dev` → http://localhost:5173; Vite proxies `/api` to `http://127.0.0.1:4000`.

---

## 6. Running migrations

**Migrations run automatically** when the Go API starts: `golang-migrate` applies embedded SQL from `backend/migrations/` before the pool serves traffic. **No separate migrate command** is required for the Docker Compose stack.

---

## 7. Test credentials

After a **fresh** database (e.g. first `docker compose up` on a **new** volume), CSV seed creates:

| Field | Value |
|-------|--------|
| **Email** | `test@example.com` |
| **Password** | `password123` |

You should see **one** project (**Demo Project**) and **three** tasks with statuses **todo**, **in_progress**, and **done**.

If an older seed already ran, run `docker compose down -v` then `docker compose up --build`, or manually clear application tables and delete the `static_csv_demo_v1` row from `app_seed_state`, then restart the API.

---

## 8. API reference

**Base URL (browser / nginx):** `http://localhost:3000/api`  
**Direct API (host-only dev):** `http://localhost:4000/api`

**Auth:** Protected routes expect `Authorization: Bearer <jwt>`. Login and register responses include `token` and `user`.

**Common errors**

- `400` — `{"error":"validation failed","fields":{"email":"is required"}}`
- `401` — `{"error":"unauthorized"}`
- `403` — `{"error":"forbidden"}`
- `404` — `{"error":"not found"}`

### Health

| Method | Path | Auth | Response |
|--------|------|------|----------|
| GET | `/healthz` | No | `{"ok":true}` |

### Auth

| Method | Path | Body | Success |
|--------|------|------|---------|
| POST | `/api/auth/register` | `{"name","email","password"}` (password ≥ 8 chars) | `201` — `{"token","user"}` |
| POST | `/api/auth/login` | `{"email","password"}` | `200` — `{"token","user"}` |
| GET | `/api/auth/me` | — | `200` — `{"user"}` |
| POST | `/api/auth/logout` | — | `204` |

**Example**

```bash
curl -s -X POST http://localhost:3000/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@example.com","password":"password123"}'
```

### Users

| Method | Path | Auth | Success |
|--------|------|------|---------|
| GET | `/api/users` | Bearer | `200` — `{"users":[...]}` (for assignee picker) |

### Projects

| Method | Path | Query | Success |
|--------|------|-------|---------|
| GET | `/api/projects` | `?page=&limit=` (default limit 20, max 100) | `200` — `projects`, `total`, `page` |
| POST | `/api/projects` | JSON body | `201` — project |
| GET | `/api/projects/{id}` | — | `200` — project + `tasks` |
| PATCH | `/api/projects/{id}` | `name`, `description`, `owner_id` (optional) | `200` |
| DELETE | `/api/projects/{id}` | — | `204` |
| GET | `/api/projects/{id}/collaborators` | — | `200` — `{"users"}` |
| GET | `/api/projects/{id}/stats` | — | `200` — `by_status`, `by_assignee` |

**Create project**

```http
POST /api/projects
Content-Type: application/json

{"name":"My project","description":"optional"}
```

**Project object (shape)**

```json
{
  "id": "uuid",
  "name": "string",
  "description": "string or omitted",
  "owner_id": "uuid",
  "created_at": "RFC3339Nano"
}
```

### Tasks

| Method | Path | Query / body | Success |
|--------|------|--------------|---------|
| GET | `/api/projects/{id}/tasks` | `?status=`, `?assignee=`, `?page=&limit=` | `200` — `tasks`, `total` |
| POST | `/api/projects/{id}/tasks` | JSON body | `201` — task |
| PATCH | `/api/tasks/{id}` | partial JSON | `200` — task |
| DELETE | `/api/tasks/{id}` | — | `204` |

**Create task (body)**

```json
{
  "title": "required",
  "description": "optional",
  "status": "todo | in_progress | done",
  "priority": "low | medium | high",
  "assignee_id": "uuid or null",
  "due_date": "YYYY-MM-DD"
}
```

**Task object (shape)**

```json
{
  "id": "uuid",
  "title": "string",
  "description": "string or omitted",
  "status": "todo|in_progress|done",
  "priority": "low|medium|high",
  "project_id": "uuid (may be omitted when nested under project detail)",
  "assignee_id": "uuid or null",
  "creator_id": "uuid",
  "due_date": "YYYY-MM-DD or null",
  "created_at": "RFC3339",
  "updated_at": "RFC3339"
}
```

**PATCH notes:** Send only fields that change. For `assignee_id`, JSON `null` or `""` clears the assignee; omit the key to leave it unchanged. `creator_id` may be updated only by the project owner.

There is **no** Postman or Bruno collection checked into this repository; copy the examples above into your HTTP client if needed.

---

## 9. What you would do with more time

- **End-to-end tests** (Playwright or Cypress) against `docker compose up` for login → project → task flows.
- **OpenAPI** document and a small **Bruno** or **Postman** export committed next to the code.
- **Production hardening:** rate limits on login, structured audit log for mutations, stricter default log level, optional switch to disable CSV seed.
- **Realtime** updates for task boards (SSE or WebSocket) instead of manual refresh.
- **Accessibility and i18n** pass on the SPA (focus management in dialogs, reduced motion, translations).
- **Frontend performance:** route-level code splitting; the production JS bundle is large today.

**Honest shortcuts:** Reviewer setup assumes Docker; seed is “run once” via a marker table, not a full idempotent reconciliation; some UI hint text may not always match the latest CSV email unless updated together.

Replace **`YOUR_GITHUB_USERNAME`** in clone URLs with your GitHub user or organization (see [§10](#10-repository-github-and-secrets)).
