APP_NAME := gophermart
CMD_DIR := .

COMPOSE_LOCAL := docker compose -f docker-compose.local.yml

GOLANGCI_IMAGE := gophermart-golangci

.PHONY: all build run test lint tidy clean sqlc accrual docker-local docker-local-rebuild lint-docker

sqlc:
	sqlc generate -f sqlc/sqlc.yaml
build:
	go build $(CMD_DIR)/cmd/gophermart/main.go
run:
	go run $(CMD_DIR)/cmd/gophermart/main.go
accrual:
	./cmd/accrual/accrual_windows_amd64.exe
lint:
	go vet ./...

# golangci-lint in Docker (mounts repo; uses docker/golangci-lint/Dockerfile).
lint-docker:
	docker build -f docker/golangci-lint/Dockerfile -t $(GOLANGCI_IMAGE) docker/golangci-lint
	docker run --rm -v "$(CURDIR):/src" -w /src $(GOLANGCI_IMAGE)
tidy:
	go mod tidy
# Start stack; rebuild app images when Dockerfiles / context change (uses layer cache).
docker-local:
	$(COMPOSE_LOCAL) up -d --build

# Recompile Go inside Docker from scratch (no cache), then start — use when you want a clean image every time.
docker-local-rebuild:
	$(COMPOSE_LOCAL) build --no-cache gophermart accrual
	$(COMPOSE_LOCAL) up -d

tests:
	go test ./...