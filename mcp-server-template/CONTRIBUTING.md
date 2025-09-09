# Contributing to MCP Server Template

Thank you for your interest in contributing to the MCP Server Template! This document provides guidelines for development, testing, and extending the template.

## Table of Contents

- [Development Environment](#development-environment)
- [Architecture Overview](#architecture-overview)
- [Adding New Features](#adding-new-features)
- [Testing Guidelines](#testing-guidelines)
- [Code Style](#code-style)
- [Extensibility Patterns](#extensibility-patterns)
- [Deployment](#deployment)

## Development Environment

### Prerequisites

- Go 1.21 or later
- Docker (for containerization)
- Git

### Setup

```bash
# Clone the repository
git clone <repository-url>
cd mcp-server-template

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Run with example config
go run cmd/server/main.go --config examples/simple-server.json --log-level debug
```

### Environment Variables

Create a `.env` file for development:

```bash
# API Keys for testing
OPENWEATHER_API_KEY=your_openweather_api_key
TEST_API_KEY=test_key_for_development

# Development settings
LOG_LEVEL=debug
ENVIRONMENT=development
```

## Architecture Overview

The MCP server template follows a modular architecture:

```
├── cmd/server/          # Application entrypoint
├── internal/
│   ├── config/         # Configuration management
│   ├── handlers/       # HTTP client and tool execution
│   ├── server/         # MCP server implementation
│   └── validation/     # Input/output validation
├── pkg/                # Public packages (future extensibility)
├── examples/           # Example configurations
└── tests/              # Test suites
```

### Key Components

1. **Configuration System** (`internal/config/`)
   - JSON-driven configuration loading
   - Environment variable substitution
   - Comprehensive validation

2. **Tool Handlers** (`internal/handlers/`)
   - Dynamic tool registration
   - HTTP client with auth support
   - Response processing and validation

3. **MCP Integration** (`internal/server/`)
   - mark3labs/mcp-go SDK integration
   - Tool, prompt, and resource management
   - HTTP server with health checks

4. **Validation** (`internal/validation/`)
   - Input parameter validation
   - Response validation
   - Security sanitization

## Adding New Features

### Adding a New Tool Type

1. **Extend Tool Configuration**:

```go
// In internal/config/types.go
type ToolConfig struct {
    // ... existing fields ...
    NewField string `json:"new_field" validate:"required"`
}
```

2. **Update Validation**:

```go
// In internal/config/loader.go
func validateBusinessRules(cfg *Config) error {
    // Add validation for new field
    for _, tool := range cfg.Tools {
        if tool.NewField == "" {
            return fmt.Errorf("new_field is required for tool %s", tool.Name)
        }
    }
    return nil
}
```

3. **Extend Tool Handler**:

```go
// In internal/handlers/tool_handler.go
func (h *ToolHandler) handleNewFeature(tool *config.ToolConfig) error {
    // Implementation for new feature
    return nil
}
```

4. **Add Tests**:

```go
// In tests/new_feature_test.go
func TestNewFeature(t *testing.T) {
    // Test implementation
}
```

### Adding New Authentication Types

1. **Extend AuthConfig**:

```go
type AuthConfig struct {
    Type string `json:"type" validate:"required,oneof=bearer basic api_key custom oauth2"`
    // Add new auth fields
    OAuth2Config *OAuth2Config `json:"oauth2_config,omitempty"`
}

type OAuth2Config struct {
    ClientID     string `json:"client_id"`
    ClientSecret string `json:"client_secret"`
    TokenURL     string `json:"token_url"`
}
```

2. **Implement in HTTP Client**:

```go
// In internal/handlers/http_client.go
func (h *HTTPClient) applyAuthentication(req *http.Request, auth *config.AuthConfig) error {
    switch auth.Type {
    // ... existing cases ...
    case "oauth2":
        return h.applyOAuth2Auth(req, auth.OAuth2Config)
    }
    return nil
}
```

### Adding New Validation Rules

1. **Register Custom Validators**:

```go
// In internal/validation/validator.go
func New() *Validator {
    validate := validator.New()
    validate.RegisterValidation("custom_rule", validateCustomRule)
    // ...
}

func validateCustomRule(fl validator.FieldLevel) bool {
    // Custom validation logic
    return true
}
```

## Testing Guidelines

### Test Structure

- **Unit Tests**: Test individual functions and components
- **Integration Tests**: Test component interactions
- **End-to-End Tests**: Test complete workflows

### Writing Tests

```go
func TestToolExecution(t *testing.T) {
    // Arrange
    tool := &config.ToolConfig{
        Name:        "test_tool",
        Endpoint:    "https://api.example.com/test",
        Method:      "GET",
    }
    
    // Act
    result, err := handler.ExecuteTool(context.Background(), "test_tool", map[string]interface{}{})
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.False(t, result.IsError)
}
```

### Mock Testing

For testing HTTP interactions:

```go
// Create test server
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}))
defer server.Close()

// Use test server URL in tool config
tool.Endpoint = server.URL
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test package
go test ./internal/config

# Run with verbose output
go test -v ./...

# Run benchmarks
go test -bench=. ./...
```

## Code Style

### Go Standards

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Use `golint` for linting
- Use `go vet` for static analysis

### Naming Conventions

- **Packages**: lowercase, single word
- **Functions**: camelCase, start with capital for exported
- **Variables**: camelCase
- **Constants**: UPPER_CASE or camelCase
- **Interfaces**: noun or adjective + "er" suffix

### Error Handling

```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to process request: %w", err)
}

// Use specific error types for different scenarios
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for field %s: %s", e.Field, e.Message)
}
```

### Logging

```go
// Use structured logging
logger.WithFields(logrus.Fields{
    "tool_name": toolName,
    "duration":  duration,
}).Info("Tool executed successfully")

// Log errors with context
logger.WithError(err).WithField("tool_name", toolName).Error("Tool execution failed")
```

## Extensibility Patterns

### Plugin Architecture

The template is designed for extensibility:

1. **Interface-based Design**: Core functionality uses interfaces
2. **Configuration-driven**: Behavior controlled by JSON config
3. **Middleware Support**: HTTP client supports middleware
4. **Event Hooks**: Lifecycle events for custom logic

### Custom Tool Types

```go
// Define interface for custom tool types
type CustomToolHandler interface {
    ExecuteTool(ctx context.Context, config *ToolConfig, params map[string]interface{}) (*APIResponse, error)
    ValidateConfig(config *ToolConfig) error
}

// Register custom handlers
type ToolRegistry struct {
    handlers map[string]CustomToolHandler
}

func (r *ToolRegistry) Register(toolType string, handler CustomToolHandler) {
    r.handlers[toolType] = handler
}
```

### Middleware Support

```go
// HTTP middleware interface
type Middleware func(http.RoundTripper) http.RoundTripper

// Example: Retry middleware
func RetryMiddleware(maxRetries int) Middleware {
    return func(next http.RoundTripper) http.RoundTripper {
        return &retryTransport{
            next:       next,
            maxRetries: maxRetries,
        }
    }
}
```

## Deployment

### Docker

```bash
# Build image
docker build -t mcp-server:latest .

# Run container
docker run -p 8080:8080 -v $(pwd)/config.json:/app/config.json mcp-server:latest
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mcp-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: mcp-server
  template:
    metadata:
      labels:
        app: mcp-server
    spec:
      containers:
      - name: mcp-server
        image: mcp-server:latest
        ports:
        - containerPort: 8080
        env:
        - name: LOG_LEVEL
          value: "info"
        volumeMounts:
        - name: config
          mountPath: /app/config.json
          subPath: config.json
      volumes:
      - name: config
        configMap:
          name: mcp-server-config
```

### Health Checks

The server provides health check endpoints:

- `/health`: Basic health status
- `/metrics`: Prometheus-compatible metrics (if enabled)

### Configuration Management

- Use ConfigMaps for configuration in Kubernetes
- Use Secrets for sensitive data (API keys)
- Support environment variable substitution

## Security Considerations

### Input Validation

- All user inputs are validated using struct tags
- Custom validation functions for complex rules
- Input sanitization to prevent injection attacks

### Authentication

- Support for multiple auth types
- Secure storage of credentials via environment variables
- Token rotation support

### Network Security

- TLS support for HTTPS endpoints
- CORS configuration
- Rate limiting

## Performance Guidelines

### Optimization

- Use connection pooling for HTTP clients
- Implement caching where appropriate
- Monitor memory usage and garbage collection
- Use context for request cancellation

### Monitoring

- Structured logging with correlation IDs
- Metrics collection (response times, error rates)
- Health check endpoints
- Graceful shutdown handling

## Contributing Process

1. **Fork the repository**
2. **Create a feature branch**
3. **Write tests for your changes**
4. **Implement the feature**
5. **Ensure all tests pass**
6. **Update documentation**
7. **Submit a pull request**

### Pull Request Guidelines

- Clear description of changes
- Reference any related issues
- Include tests for new functionality
- Update documentation as needed
- Follow code style guidelines

Thank you for contributing to the MCP Server Template! 