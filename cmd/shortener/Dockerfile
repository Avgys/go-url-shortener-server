# Build from repository root:
#   docker build -f cmd/shortener/Dockerfile -t url-shortener:latest .
#   docker build -f cmd/shortener/Dockerfile --build-arg VERSION=1.0.0 --build-arg BUILD_DATE=2026-06-11 -t url-shortener:latest .

FROM golang:1.26-alpine AS builder

ARG VERSION=1.0.0
ARG BUILD_DATE

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN set -eux; \
    if [ -z "${BUILD_DATE:-}" ]; then BUILD_DATE="$(date +%Y-%m-%d)"; fi; \
    CGO_ENABLED=0 GOOS=linux go build -trimpath \
      -ldflags "-s -w -X main.BuildVersion=${VERSION} -X main.BuildDate=${BUILD_DATE}" \
      -o ./shortener ./cmd/shortener

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /app/shortener /app/shortener
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080

ENTRYPOINT ["/app/shortener"]
