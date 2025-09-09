SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := help

GO := go
BIN_DIR := bin
BACKEND_BIN := $(BIN_DIR)/backend
SERVER_BIN := $(BIN_DIR)/backend-server

PKG := mcp-backend
CMD_BACKEND := ./cmd/backend
CMD_SERVER := ./cmd/server

PORT ?= 6161

IMAGE_REPO ?= yourrepo/mcp-backend
IMAGE_TAG ?= latest

NAMESPACE ?= mcp
RELEASE ?= mcp-backend

GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_TAG := $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.0.0")
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

LDFLAGS := -s -w

## Show this help
help:
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_.-]+:.*##/ {printf "\033[36m%-24s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## Download go module dependencies
deps:
	$(GO) mod download

## Tidy go.mod / go.sum
tidy:
	$(GO) mod tidy

## Format source code
fmt:
	$(GO) fmt ./...

## Vet static checks
vet:
	$(GO) vet ./...

## Lint with go vet and optional golangci-lint if installed
lint: vet
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed; skipping"

## Run tests with race detector and coverage
test:
	$(GO) test -race -coverprofile=coverage.out ./...

## Build backend CLI binary
build: build-backend ## Build the primary backend binary

## Build all binaries (CLI and legacy server)
build-all: build-backend build-server

## Build backend CLI
build-backend: $(BACKEND_BIN)

$(BACKEND_BIN): ensure-bin
	$(GO) build -trimpath -ldflags '$(LDFLAGS)' -o $@ $(CMD_BACKEND)

## Build legacy server binary
build-server: $(SERVER_BIN)

$(SERVER_BIN): ensure-bin
	$(GO) build -trimpath -ldflags '$(LDFLAGS)' -o $@ $(CMD_SERVER)

## Run backend locally (PORT?=6161)
run: build-backend
	$(BACKEND_BIN) server --port $(PORT)

## Run backend in dev mode (no build)
run-dev:
	$(GO) run $(CMD_BACKEND) server --port $(PORT)

## Build Docker image
docker-build:
	docker build -t $(IMAGE_REPO):$(IMAGE_TAG) .

## Push Docker image
docker-push:
	docker push $(IMAGE_REPO):$(IMAGE_TAG)

## Run Docker container mapping port
docker-run:
	docker run --rm -p $(PORT):6161 --name backend $(IMAGE_REPO):$(IMAGE_TAG) /backend server --port 6161

## Apply k8s manifests (direct)
k8s-apply:
	kubectl apply -f k8s.yaml

## Install/upgrade backend via Helm chart
helm-install:
	helm upgrade --install $(RELEASE) deploy/helm -n $(NAMESPACE) --create-namespace \
		--set image.repository=$(IMAGE_REPO) --set image.tag=$(IMAGE_TAG)

## Uninstall backend Helm release
helm-uninstall:
	helm uninstall $(RELEASE) -n $(NAMESPACE)

## Clean build artifacts
clean:
	rm -rf $(BIN_DIR) coverage.out backend backend-server

.PHONY: help deps tidy fmt vet lint test build build-all build-backend build-server run run-dev docker-build docker-push docker-run k8s-apply helm-install helm-uninstall clean ensure-bin

ensure-bin:
	mkdir -p $(BIN_DIR)


