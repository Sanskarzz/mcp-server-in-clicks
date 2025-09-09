# Multi-stage Dockerfile for MCP server template
FROM golang:1.23-alpine AS builder

# Set necessary environment variables for building
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./

    # Download dependencies
    ENV GOPROXY=direct
    RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -ldflags="-w -s" -o mcp-server cmd/server/main.go

# Create minimal runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/mcp-server .

# Copy example configurations (optional)
COPY --from=builder /app/examples ./examples

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port (configurable via environment)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Set default command
ENTRYPOINT ["./mcp-server"]
CMD ["--config", "config.json", "--port", "8080"] 