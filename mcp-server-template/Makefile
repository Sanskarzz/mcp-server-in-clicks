# Makefile for MCP Server Template

# Application info
APP_NAME := mcp-server-template
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date +%Y-%m-%dT%H:%M:%S%z)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go related variables
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := gofmt
GOLINT := golangci-lint
BINARY_NAME := $(APP_NAME)
BINARY_PATH := ./$(BINARY_NAME)

# Build flags
LDFLAGS := -ldflags "-w -s -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

# Docker related variables
DOCKER_REGISTRY := sanskardevops
DOCKER_IMAGE := $(DOCKER_REGISTRY)/$(APP_NAME)
DOCKER_TAG := $(VERSION)

# Kubernetes/Helm related variables
HELM_CHART := ./deploy/helm
KUBERNETES_NAMESPACE := mcp-servers

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

.PHONY: help build run test clean fmt lint vet deps docker helm

# Default target
.DEFAULT_GOAL := help

## Display this help message
help:
	@echo "$(GREEN)MCP Server Template Makefile$(NC)"
	@echo "$(YELLOW)Available targets:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(BLUE)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## Build the application
build:
	@echo "$(GREEN)Building $(APP_NAME)...$(NC)"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH) cmd/server/main.go
	@echo "$(GREEN)Build complete: $(BINARY_PATH)$(NC)"

## Build for multiple platforms
build-all:
	@echo "$(GREEN)Building for multiple platforms...$(NC)"
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 cmd/server/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 cmd/server/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 cmd/server/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 cmd/server/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe cmd/server/main.go
	@echo "$(GREEN)Multi-platform build complete$(NC)"

## Run the application with simple config
run:
	@echo "$(GREEN)Running $(APP_NAME) with simple config...$(NC)"
	$(GOCMD) run cmd/server/main.go --config examples/simple-server.json --log-level debug

## Run the application with weather config
run-weather:
	@echo "$(GREEN)Running $(APP_NAME) with weather config...$(NC)"
	$(GOCMD) run cmd/server/main.go --config examples/weather-server.json --log-level debug

## Run all tests
test:
	@echo "$(GREEN)Running tests...$(NC)"
	$(GOTEST) -v ./...

## Run tests with coverage
test-coverage:
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

## Run tests with race detection
test-race:
	@echo "$(GREEN)Running tests with race detection...$(NC)"
	$(GOTEST) -v -race ./...

## Run benchmarks
bench:
	@echo "$(GREEN)Running benchmarks...$(NC)"
	$(GOTEST) -bench=. -benchmem ./...

## Clean build artifacts
clean:
	@echo "$(GREEN)Cleaning...$(NC)"
	$(GOCLEAN)
	rm -f $(BINARY_PATH)
	rm -rf dist/
	rm -f coverage.out coverage.html

## Format Go code
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	$(GOFMT) -w .

## Run Go vet
vet:
	@echo "$(GREEN)Running go vet...$(NC)"
	$(GOCMD) vet ./...

## Run linter
lint:
	@echo "$(GREEN)Running linter...$(NC)"
	$(GOLINT) run ./...

## Install/update dependencies
deps:
	@echo "$(GREEN)Installing dependencies...$(NC)"
	$(GOMOD) download
	$(GOMOD) tidy

## Update dependencies
deps-update:
	@echo "$(GREEN)Updating dependencies...$(NC)"
	$(GOMOD) get -u ./...
	$(GOMOD) tidy

## Run all quality checks
check: fmt vet lint test
	@echo "$(GREEN)All quality checks passed!$(NC)"

## Build Docker image
docker-build:
	@echo "$(GREEN)Building Docker image...$(NC)"
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest
	@echo "$(GREEN)Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)$(NC)"

## Push Docker image
docker-push: docker-build
	@echo "$(GREEN)Pushing Docker image...$(NC)"
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest
	@echo "$(GREEN)Docker image pushed$(NC)"

## Run Docker container locally
docker-run:
	@echo "$(GREEN)Running Docker container...$(NC)"
	docker run -it --rm -p 8080:8080 \
		-v $(PWD)/examples/simple-server.json:/app/config.json \
		$(DOCKER_IMAGE):latest

## Run Docker container with weather config
docker-run-weather:
	@echo "$(GREEN)Running Docker container with weather config...$(NC)"
	docker run -it --rm -p 8080:8080 \
		-v $(PWD)/examples/weather-server.json:/app/config.json \
		-e OPENWEATHER_API_KEY=${OPENWEATHER_API_KEY} \
		$(DOCKER_IMAGE):latest

## Install Helm chart
helm-install:
	@echo "$(GREEN)Installing Helm chart...$(NC)"
	helm upgrade --install $(APP_NAME) $(HELM_CHART) \
		--namespace $(KUBERNETES_NAMESPACE) \
		--create-namespace \
		--set image.tag=$(DOCKER_TAG)

## Uninstall Helm chart
helm-uninstall:
	@echo "$(GREEN)Uninstalling Helm chart...$(NC)"
	helm uninstall $(APP_NAME) --namespace $(KUBERNETES_NAMESPACE)

## Package Helm chart
helm-package:
	@echo "$(GREEN)Packaging Helm chart...$(NC)"
	helm package $(HELM_CHART) --destination ./dist/

## Lint Helm chart
helm-lint:
	@echo "$(GREEN)Linting Helm chart...$(NC)"
	helm lint $(HELM_CHART)

## Template Helm chart (dry-run)
helm-template:
	@echo "$(GREEN)Templating Helm chart...$(NC)"
	helm template $(APP_NAME) $(HELM_CHART) \
		--namespace $(KUBERNETES_NAMESPACE) \
		--set image.tag=$(DOCKER_TAG)

## Generate API documentation
docs:
	@echo "$(GREEN)Generating documentation...$(NC)"
	$(GOCMD) doc -all > docs/api.txt
	@echo "$(GREEN)Documentation generated: docs/api.txt$(NC)"

## Run security scan
security:
	@echo "$(GREEN)Running security scan...$(NC)"
	gosec ./...

## Setup development environment
dev-setup:
	@echo "$(GREEN)Setting up development environment...$(NC)"
	$(GOMOD) download
	@echo "Installing development tools..."
	$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOCMD) install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@echo "$(GREEN)Development environment setup complete$(NC)"

## Create a new release
release: check build-all docker-build helm-package
	@echo "$(GREEN)Creating release $(VERSION)...$(NC)"
	@echo "$(YELLOW)Built binaries in dist/$(NC)"
	@echo "$(YELLOW)Docker image: $(DOCKER_IMAGE):$(DOCKER_TAG)$(NC)"
	@echo "$(YELLOW)Helm chart packaged$(NC)"
	@echo "$(GREEN)Release $(VERSION) ready!$(NC)"

## Show current version
version:
	@echo "$(GREEN)Version: $(VERSION)$(NC)"
	@echo "$(GREEN)Git Commit: $(GIT_COMMIT)$(NC)"
	@echo "$(GREEN)Build Time: $(BUILD_TIME)$(NC)"

## Watch for changes and rebuild
watch:
	@echo "$(GREEN)Watching for changes...$(NC)"
	@while true; do \
		inotifywait -e modify,create,delete -r . --exclude '\.git|\.tmp|dist|coverage' -q && \
		make build && \
		echo "$(GREEN)Rebuilt at $$(date)$(NC)"; \
	done

## Benchmark comparison
bench-compare:
	@echo "$(GREEN)Running benchmark comparison...$(NC)"
	@if [ ! -f bench-old.txt ]; then \
		$(GOTEST) -bench=. -benchmem ./... > bench-old.txt; \
		echo "$(YELLOW)Baseline benchmark saved. Run again to compare.$(NC)"; \
	else \
		$(GOTEST) -bench=. -benchmem ./... > bench-new.txt; \
		benchcmp bench-old.txt bench-new.txt; \
		mv bench-new.txt bench-old.txt; \
	fi

## Load test the server
load-test:
	@echo "$(GREEN)Running load test...$(NC)"
	@if command -v hey >/dev/null 2>&1; then \
		hey -n 1000 -c 10 http://localhost:8080/health; \
	else \
		echo "$(RED)Please install 'hey' for load testing: go install github.com/rakyll/hey@latest$(NC)"; \
	fi

## Show project statistics
stats:
	@echo "$(GREEN)Project Statistics:$(NC)"
	@echo "Lines of Go code: $$(find . -name '*.go' -not -path './vendor/*' | xargs wc -l | tail -1 | awk '{print $$1}')"
	@echo "Number of Go files: $$(find . -name '*.go' -not -path './vendor/*' | wc -l)"
	@echo "Number of tests: $$(grep -r 'func Test' --include='*.go' . | wc -l)"
	@echo "Dependencies: $$(go list -m all | wc -l)"

# Include local Makefile for custom targets
-include Makefile.local 