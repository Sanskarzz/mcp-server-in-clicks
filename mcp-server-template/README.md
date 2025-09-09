# MCP Server Template

A modular, config-driven Go template for creating custom Model Context Protocol (MCP) servers using the [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) SDK.

## Overview

This template enables dynamic MCP server creation without code changes. All tools, prompts, and resources are defined through JSON configuration, making it perfect for no-code SaaS platforms.

## Features

- 🔧 **Dynamic Tool Registration**: Define API endpoint tools via JSON config
- 🔐 **Authentication Support**: Configurable auth headers and token handling
- ✅ **Input/Output Validation**: Robust type validation and error handling
- 🧪 **Testing Framework**: Comprehensive test scaffolding
- 📝 **Prompts & Resources**: Configurable prompts and static resources
- 🚀 **Production Ready**: Designed for Kubernetes deployment

## Quick Start

```bash
# Clone and setup
cd mcp-server-template
go mod tidy

# Run with example config
go run cmd/server/main.go --config examples/weather-server.json

# Run tests
go test ./...
```

## Configuration

The server is entirely configured through JSON:

```json
{
  "server": {
    "name": "weather-api-server",
    "version": "1.0.0",
    "description": "Weather data MCP server"
  },
  "tools": [
    {
      "name": "get_weather",
      "description": "Get current weather for a location",
      "endpoint": "https://api.openweathermap.org/data/2.5/weather",
      "method": "GET",
      "headers": {
        "Authorization": "Bearer ${API_KEY}"
      },
      "parameters": [
        {
          "name": "location",
          "type": "string",
          "description": "City name",
          "required": true
        }
      ]
    }
  ]
}
```

## Architecture

```
mcp-server-template/
├── cmd/server/           # Main application entrypoint
├── internal/
│   ├── config/          # Configuration loading and validation
│   ├── handlers/        # Tool handlers and HTTP client
│   ├── server/          # MCP server implementation
│   └── validation/      # Input/output validation
├── pkg/                 # Public packages
├── examples/            # Example configurations
└── tests/               # Test suites
```

## Deployment

Designed for containerized deployment in Kubernetes:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o mcp-server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/mcp-server /usr/local/bin/
ENTRYPOINT ["mcp-server"]
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines and extensibility patterns. 