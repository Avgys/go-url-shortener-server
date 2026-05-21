APP_NAME := shortener
CMD_DIR := .

export CGO_ENABLED := 1

COMPOSE_LOCAL := docker compose -f docker-compose.local.yml

GOLANGCI_IMAGE := ${APP_NAME}-golangci

.PHONY: all build run test lint tidy clean sqlc audit mocks docker-local docker-local-rebuild lint-docker

sqlc:
	sqlc generate -f sqlc/sqlc.yaml
mocks:
	go generate ./internal/service/audit/...
	go generate ./internal/service/shortifier/...
build:
	go build $(CMD_DIR)/cmd/${APP_NAME}/main.go
run:
	go run $(CMD_DIR)/cmd/${APP_NAME}/main.go
audit:
	./cmd/audit/audit_windows_amd64.exe
lint:
	go vet ./...

# golangci-lint in Docker (mounts repo; uses docker/golangci-lint/Dockerfile).
# Do not pass -w /src: Git Bash on Windows rewrites it to C:/Program Files/Git/src.
lint-docker:
	docker build -f docker/golangci-lint/Dockerfile -t $(GOLANGCI_IMAGE) docker/golangci-lint
	docker run --rm -v "$(CURDIR):/src" $(GOLANGCI_IMAGE)
tidy:
	go mod tidy
# Start stack; rebuild app images when Dockerfiles / context change (uses layer cache).
docker-local:
	$(COMPOSE_LOCAL) up -d --build

# Recompile Go inside Docker from scratch (no cache), then start — use when you want a clean image every time.
docker-local-rebuild:
	$(COMPOSE_LOCAL) build --no-cache ${APP_NAME}
	$(COMPOSE_LOCAL) up -d

tests:
	go test ./...

tests-coverage:
	go test ./... -coverprofile=coverage.out -race
	go tool cover -func=coverage.out