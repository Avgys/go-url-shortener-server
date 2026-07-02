# URL Shortener Server

![tests](https://img.shields.io/endpoint?url=https%3A%2F%2Fgist.githubusercontent.com%2FAvgys%2F6d76e76f4819d555ec85c6089ac087bb%2Fraw%2Fgo-url-shortener-go-tests.json)
![coverage](https://img.shields.io/endpoint?url=https%3A%2F%2Fgist.githubusercontent.com%2FAvgys%2F6d76e76f4819d555ec85c6089ac087bb%2Fraw%2Fgo-url-shortener-go-coverage.json)

HTTP service that turns long URLs into short links, resolves redirects, and stores mappings per anonymous user (JWT cookie). Supports PostgreSQL, file-backed storage, or an in-memory store for development.

---

## How it works

1. **Shorten** — Client sends a long URL (`POST /` as plain text, or JSON on `/api/shorten` / `/api/shorten/batch`). The service generates a short ID, stores the pair, and returns a full short URL built from `BASE_URL`.
2. **Redirect** — `GET /{shortURL}` looks up the original URL and responds with `307 Temporary Redirect`.
3. **User URLs** — A JWT auth cookie (`AUTH_COOKIE`) identifies the user. Middleware creates or refreshes the cookie on first visit. Authenticated routes list or soft-delete the user’s links.
4. **Storage** — If `DATABASE_DSN` is set, data lives in PostgreSQL (with [sqlc](https://sqlc.dev/) queries). Otherwise `FILE_STORAGE_PATH` selects a CSV file store; if neither is set, an in-memory store is used.
5. **Migrations** — On startup with PostgreSQL, [golang-migrate](https://github.com/golang-migrate/migrate) applies SQL from `migrations/sql`.

Middleware includes request logging, gzip request/response handling, and panic recovery.

---

## Graceful shutdown

When the process receives `SIGINT`/`SIGTERM`, the app starts graceful shutdown:

- stops accepting new HTTP requests;
- waits up to `5s` for in-flight handlers to finish;
- cancels unfinished request contexts after the timeout.

This means an already running `POST /api/shorten/batch` may still finish if it completes within the shutdown timeout.

---

## REST API

| Method | Path | Content-Type | Auth cookie | Description |
|--------|------|--------------|-------------|-------------|
| `POST` | `/` | `text/plain`, `application/x-gzip` | Set on response | Shorten URL (body = long URL) |
| `POST` | `/api/shorten` | `application/json` | Set on response | Shorten one URL (`{"url":"..."}`) |
| `POST` | `/api/shorten/batch` | `application/json` | Set on response | Batch shorten with correlation IDs |
| `GET` | `/{shortURL}` | — | — | Redirect to original URL |
| `GET` | `/ping` | — | — | Health check (DB ping when using Postgres) |
| `GET` | `/api/user/urls` | — | Required | List current user’s URL pairs |
| `DELETE` | `/api/user/urls` | `application/json` | Required | Soft-delete short URLs (JSON array of short keys) |

Batch shorten request/response shapes live in `internal/model/requests` and `internal/model/responses`.

---

## Configuration

Environment variables (parsed after CLI flags; env wins over flags). See `internal/config/config.go`.

| Variable | Flag | Meaning |
|----------|------|---------|
| `SERVER_ADDRESS` | `-a` | HTTP listen address (default `localhost:8080`) |
| `BASE_URL` | `-b` | Public base URL for generated short links (default `http://localhost:8080`) |
| `DATABASE_DSN` | `-d` | PostgreSQL connection string; omit to use file or memory |
| `FILE_STORAGE_PATH` | `-f` | Path for file storage when no DSN is set |
| `AUDIT_FILE` | `-audit-file` | Path to audit log file; empty disables file audit |
| `AUDIT_URL` | `-audit-url` | Full URL of remote audit receiver; empty disables remote audit |

Example:

```bash
export SERVER_ADDRESS=:8080
export BASE_URL=http://localhost:8080
export DATABASE_DSN=postgresql://postgres:postgres@localhost:5432/praktikum?sslmode=disable
```

---

## Project layout

```
cmd/shortener/          # Application entrypoint and Dockerfile
internal/
  config/               # Flags and env configuration
  handler/              # HTTP handlers
  router/               # Chi routes and middleware wiring
  middlewares/          # Logging, compression, auth cookie, recovery
  service/shortifier/   # Shortening, resolve, delete logic
  repository/           # In-memory, file, and PostgreSQL stores
  db/                   # Connection pool and migrations
sqlc/                   # Generated queries from SQL
migrations/sql/         # PostgreSQL migrations
docker-compose.local.yml
Makefile
```

---

## Local development

### Prerequisites

- Go 1.26+
- [Docker](https://docs.docker.com/get-docker/) (optional, for Postgres and containerized lint)

### Build and run

```bash
make build    # go build ./cmd/shortener
make run      # run the server locally
make tests    # go test ./...
make tidy     # go mod tidy
make sqlc     # regenerate sqlc code from sqlc/sqlc.yaml
```

Integration tests that need Postgres skip unless `TEST_DATABASE_DSN` is set.

### Lint

```bash
make lint         # go vet ./...
make lint-docker  # golangci-lint in Docker
```

---

## Deploy with Docker Compose

`docker-compose.local.yml` starts **PostgreSQL** and the **shortener** service on one network.

### Quick start

From the repository root:

```bash
docker compose -f docker-compose.local.yml up -d --build
```

Or:

```bash
make docker-local
```

### Services

| Service | Role | Default host port |
|---------|------|-------------------|
| `postgres` | Database | `5432` (`POSTGRES_PORT`) |
| `shortener` | This API | `8080` (`SHORTENER_PORT`) |

The shortener image is built from `cmd/shortener/Dockerfile` (build context: repo root). It waits for Postgres to be healthy before starting.

### Compose environment

| Variable | Used for |
|----------|----------|
| `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` | Postgres container |
| `RUN_ADDRESS` | Mapped to `SERVER_ADDRESS` in the app (default `:8080`) |
| `REDIRECT_DOMAIN` | Mapped to `BASE_URL` (default `http://localhost:8080`) |
| `FILE_STORAGE_PATH` | Optional file store inside the container |

`DATABASE_DSN` is composed automatically to point at the `postgres` service.

### Rebuild without cache

```bash
make docker-local-rebuild
```

### Stop

```bash
docker compose -f docker-compose.local.yml down
```

Remove the Postgres volume (destructive):

```bash
docker compose -f docker-compose.local.yml down -v
```

---

## Module

Go module: `go-url-shortener` (Go 1.26).

## Optimization log

Previously, compression ran for every response, including very small bodies. The middleware now applies gzip only when the response body is at least 1400 bytes, which reduced allocations on small responses.

Build ID: D:\GoProjects\go-url-shortener-server\cmd\shortener\__debug_bin.exe2026-05-28 18:36:21.6070308 +0300 +03
Type: alloc_space
Time: 2026-05-28 15:48:47 +03
Showing nodes accounting for -785.90MB, 97.88% of 802.91MB total
Dropped 109 nodes (cum <= 4.01MB)
      flat  flat%   sum%        cum   cum%
 -641.68MB 79.92% 79.92%  -774.88MB 96.51%  compress/flate.NewWriter
 -129.20MB 16.09% 96.01%  -129.20MB 16.09%  compress/flate.(*compressor).initDeflate
   -6.01MB  0.75% 96.76%    -6.01MB  0.75%  sync.(*Pool).pinSlow
   -5.50MB  0.69% 97.45%    -5.50MB  0.69%  github.com/rs/zerolog.Logger.With
   -3.50MB  0.44% 97.88%       -4MB   0.5%  compress/flate.newHuffmanBitWriter
         0     0% 97.88%  -133.20MB 16.59%  compress/flate.(*compressor).init
         0     0% 97.88%  -774.88MB 96.51%  compress/gzip.(*Writer).Close
         0     0% 97.88%  -774.88MB 96.51%  compress/gzip.(*Writer).Write
         0     0% 97.88%  -787.89MB 98.13%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 97.88%  -784.89MB 97.76%  github.com/go-chi/chi/v5/middleware.RealIP.func1
         0     0% 97.88%  -784.89MB 97.76%  go-url-shortener/internal/middlewares.Recoverer.func1
         0     0% 97.88%  -776.38MB 96.70%  go-url-shortener/internal/middlewares.WithCompression.func1
         0     0% 97.88%  -774.88MB 96.51%  go-url-shortener/internal/middlewares.WithCompression.func1.2
         0     0% 97.88%  -784.89MB 97.76%  go-url-shortener/internal/middlewares.WithLogging.func1
         0     0% 97.88%  -774.88MB 96.51%  go-url-shortener/internal/middlewares/compress.(*compressWriter).Close
         0     0% 97.88%  -790.40MB 98.44%  net/http.(*conn).serve
         0     0% 97.88%  -784.89MB 97.76%  net/http.HandlerFunc.ServeHTTP
         0     0% 97.88%  -787.89MB 98.13%  net/http.serverHandler.ServeHTTP
         0     0% 97.88%    -6.01MB  0.75%  sync.(*Pool).pin
