# Gophermart (loyalty service)

![tests](https://img.shields.io/endpoint?url=https%3A%2F%2Fgist.githubusercontent.com%2FAvgys%2F6d76e76f4819d555ec85c6089ac087bb%2Fraw%2Fgo-gophermat-course-go-tests.json)
![coverage](https://img.shields.io/endpoint?url=https%3A%2F%2Fgist.githubusercontent.com%2FAvgys%2F6d76e76f4819d555ec85c6089ac087bb%2Fraw%2Fgo-gophermat-course-go-coverage.json)

HTTP service for a **loyalty program**: users earn points (scores) on qualifying orders and **spend those points on later purchases** via withdrawals. Most behavior is exposed through a **simple JSON REST API** (registration, login, orders, balance, withdrawals).

---

## How it works

1. **Client → API** — The user submits an order number to accrue points (`POST /api/user/orders`). The order is validated and **persisted in PostgreSQL**.
2. **Background polling** — A routine periodically loads orders that still need accrual processing.
3. **Worker pool** — A **goroutine pool** polls the external **accrual** service for each order: **status** and **accrued score**. Results are applied so the balance reflects earned points.
4. **Spending points** — The user checks balance and records withdrawals against future purchases (`GET /api/user/balance`, `POST /api/user/balance/withdraw`, `GET /api/user/withdrawals`).

Database **migrations** run automatically when the app starts (`golang-migrate`, `migrations/sql`).

---

## REST API (overview)

| Area | Method | Path | Notes |
|------|--------|------|--------|
| Auth | `POST` | `/api/user/register` | JSON body |
| Auth | `POST` | `/api/user/login` | JSON body |
| Orders | `POST` | `/api/user/orders` | Submit order number (cookie auth) |
| Orders | `GET` | `/api/user/orders` | List user orders |
| Balance | `GET` | `/api/user/balance` | Current balance |
| Balance | `POST` | `/api/user/balance/withdraw` | Withdraw points (JSON) |
| Withdrawals | `GET` | `/api/user/withdrawals` | Withdrawal history |

Protected routes use session cookie middleware after login.

---

## Configuration

Environment variables (see `internal/config/config.go`):

| Variable | Meaning |
|----------|---------|
| `RUN_ADDRESS` | HTTP listen address (e.g. `:8080`) |
| `DATABASE_URI` | PostgreSQL connection string |
| `ACCRUAL_SYSTEM_ADDRESS` | Base URL of the accrual service (e.g. `http://accrual:8080` in Docker) |

Optional CLI overrides: `-a` (address), `-d` (database URI), `-r` (accrual URL).

---

## Deploy with Docker Compose

The repo ships with **`docker-compose.local.yml`**: **PostgreSQL**, the **accrual** service, and **gophermart** on one network.

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose v2

### Quick start

From the repository root:

```bash
docker compose -f docker-compose.local.yml up -d --build
```

Or use the Makefile helper (same command):

```bash
make docker-local
```

### What gets started

| Service | Role | Default host port |
|---------|------|-------------------|
| `postgres` | Database | `5432` (override with `POSTGRES_PORT`) |
| `accrual` | Accrual API used by the worker pool | `8081` (`ACCRUAL_PORT`, maps to container `8080`) |
| `gophermart` | This REST API | `8080` (`GOPHERMART_PORT`) |

`gophermart` waits for Postgres to be healthy, then starts after `accrual` is up. Images are built from:

- `cmd/gophermart/Dockerfile` — main API
- `cmd/accrual/Dockerfile` — accrual service

### Environment overrides

Compose uses variables with defaults (see `docker-compose.local.yml`), for example:

- `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`
- `RUN_ADDRESS` (gophermart listen address inside the container)
- `ACCRUAL_SYSTEM_ADDRESS` — should point at the accrual service from gophermart (default `http://accrual:8080`)

### Rebuild without cache

```bash
make docker-local-rebuild
```

### Stop and remove containers

```bash
docker compose -f docker-compose.local.yml down
```

To remove the Postgres volume as well (destructive):

```bash
docker compose -f docker-compose.local.yml down -v
```

---

## Local development (without Docker)

- `make build` / `make run` — build or run `cmd/gophermart`
- `make tests` — run `go test ./...` (see `Makefile` for other targets)

Ensure Postgres is reachable and `DATABASE_URI` / `ACCRUAL_SYSTEM_ADDRESS` are set for your environment.

---

## Module

Go module: `avgys-gophermat` (Go 1.26).
