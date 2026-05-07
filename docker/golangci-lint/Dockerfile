# Build (small context — only this folder):
#   docker build -f docker/golangci-lint/Dockerfile -t gophermart-golangci docker/golangci-lint
#
# Run from repository root:
#   docker run --rm -v "$PWD:/src" -w /src gophermart-golangci

FROM golang:1.26-alpine

RUN apk add --no-cache git ca-certificates

ARG GOLANGCI_LINT_VERSION=v2.8.0
ENV CGO_ENABLED=0
# Avoid "error obtaining VCS status" when .git is missing or shallow (e.g. some CI checkouts).
ENV GOFLAGS=-buildvcs=false

RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}

ENV PATH="/go/bin:${PATH}"

WORKDIR /src

ENTRYPOINT ["golangci-lint"]
CMD ["run", "./..."]
