package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mcp-server-template/internal/config"
	"mcp-server-template/internal/handlers"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
)

// MCPServer wraps the mark3labs MCP server with our configuration-driven logic
type MCPServer struct {
	mcpServer   *server.MCPServer
	config      *config.Config
	toolHandler *handlers.ToolHandler
	logger      *logrus.Logger
	httpServer  *http.Server
}

// New creates a new configured MCP server instance
func New(cfg *config.Config) (*MCPServer, error) {
	logger := logrus.New()

	// Configure logging
	level, err := logrus.ParseLevel(cfg.Runtime.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	if cfg.Runtime.Environment == "production" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	logger.WithField("server_name", cfg.Server.Name).Info("Creating MCP server")

	// Create the underlying MCP server with capabilities
	mcpServer := server.NewMCPServer(
		cfg.Server.Name,
		cfg.Server.Version,
		server.WithPromptCapabilities(true),
		server.WithResourceCapabilities(true, true),
	)

	// Create tool handler
	toolHandler := handlers.NewToolHandler()

	// Create our wrapper
	mcpServerWrapper := &MCPServer{
		mcpServer:   mcpServer,
		config:      cfg,
		toolHandler: toolHandler,
		logger:      logger,
	}

	// Configure the server
	if err := mcpServerWrapper.configure(); err != nil {
		return nil, fmt.Errorf("failed to configure server: %w", err)
	}

	return mcpServerWrapper, nil
}

// configure sets up the MCP server with tools, prompts, and resources
func (s *MCPServer) configure() error {
	s.logger.Info("Configuring MCP server")

	// Register tools
	if err := s.registerTools(); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}

	// Register prompts
	if err := s.registerPrompts(); err != nil {
		return fmt.Errorf("failed to register prompts: %w", err)
	}

	// Register resources
	if err := s.registerResources(); err != nil {
		return fmt.Errorf("failed to register resources: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"tools_count":     len(s.config.Tools),
		"prompts_count":   len(s.config.Prompts),
		"resources_count": len(s.config.Resources),
	}).Info("MCP server configured successfully")

	return nil
}

// registerTools registers all configured tools
func (s *MCPServer) registerTools() error {
	if len(s.config.Tools) == 0 {
		s.logger.Info("No tools to register")
		return nil
	}

	// Register tools with the tool handler
	if err := s.toolHandler.RegisterTools(s.mcpServer, s.config.Tools); err != nil {
		return err
	}

	// Tools are now registered individually in the tool handler with their callbacks

	return nil
}

// registerPrompts registers all configured prompts
func (s *MCPServer) registerPrompts() error {
	if len(s.config.Prompts) == 0 {
		s.logger.Info("No prompts to register")
		return nil
	}

	s.logger.WithField("prompts_count", len(s.config.Prompts)).Info("Registering prompts")

	// Convert and register each prompt
	for _, promptConfig := range s.config.Prompts {
		prompt := s.convertToMCPPrompt(&promptConfig)

		// Register prompt with handler
		s.mcpServer.AddPrompt(prompt, func(arguments map[string]string) (*mcp.GetPromptResult, error) {
			// Substitute arguments in the prompt content
			content := promptConfig.Content
			for key, value := range arguments {
				placeholder := fmt.Sprintf("{%s}", key)
				content = strings.ReplaceAll(content, placeholder, value)
			}

			return mcp.NewGetPromptResult(promptConfig.Description, []mcp.PromptMessage{
				mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(content)),
			}), nil
		})

		s.logger.WithField("prompt_name", promptConfig.Name).Debug("Prompt registered")
	}

	return nil
}

// convertToMCPPrompt converts a config prompt to an MCP prompt
func (s *MCPServer) convertToMCPPrompt(promptConfig *config.PromptConfig) mcp.Prompt {
	// Build prompt options
	var opts []mcp.PromptOption
	opts = append(opts, mcp.WithPromptDescription(promptConfig.Description))

	// Add arguments
	for _, arg := range promptConfig.Arguments {
		var argOpts []mcp.ArgumentOption
		argOpts = append(argOpts, mcp.ArgumentDescription(arg.Description))
		if arg.Required {
			argOpts = append(argOpts, mcp.RequiredArgument())
		}
		opts = append(opts, mcp.WithArgument(arg.Name, argOpts...))
	}

	return mcp.NewPrompt(promptConfig.Name, opts...)
}

// registerResources registers all configured resources
func (s *MCPServer) registerResources() error {
	if len(s.config.Resources) == 0 {
		s.logger.Info("No resources to register")
		return nil
	}

	s.logger.WithField("resources_count", len(s.config.Resources)).Info("Registering resources")

	// Convert and register each resource
	for _, resourceConfig := range s.config.Resources {
		resource := s.convertToMCPResource(&resourceConfig)

		// Register resource with handler
		s.mcpServer.AddResource(resource, func(request mcp.ReadResourceRequest) ([]interface{}, error) {
			// Get resource content
			content, err := s.getResourceContent(&resourceConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to get resource content: %w", err)
			}

			return []interface{}{mcp.NewTextContent(content)}, nil
		})

		s.logger.WithField("resource_uri", resourceConfig.URI).Debug("Resource registered")
	}

	return nil
}

// convertToMCPResource converts a config resource to an MCP resource
func (s *MCPServer) convertToMCPResource(resourceConfig *config.ResourceConfig) mcp.Resource {
	var opts []mcp.ResourceOption
	opts = append(opts, mcp.WithResourceDescription(resourceConfig.Description))
	opts = append(opts, mcp.WithMIMEType(resourceConfig.MimeType))

	return mcp.NewResource(resourceConfig.URI, resourceConfig.Name, opts...)
}

// getResourceContent retrieves the content for a resource
func (s *MCPServer) getResourceContent(resource *config.ResourceConfig) (string, error) {
	// Inline content
	if resource.Content != "" {
		return resource.Content, nil
	}

	// File path content
	if resource.FilePath != "" {
		// Make path relative to current working directory if not absolute
		path := resource.FilePath
		if !filepath.IsAbs(path) {
			wd, err := os.Getwd()
			if err != nil {
				return "", fmt.Errorf("failed to get working directory: %w", err)
			}
			path = filepath.Join(wd, path)
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to read file %s: %w", path, err)
		}
		return string(content), nil
	}

	// URL content (simple HTTP GET)
	if resource.URL != "" {
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Get(resource.URL)
		if err != nil {
			return "", fmt.Errorf("failed to fetch URL %s: %w", resource.URL, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("HTTP error %d when fetching %s", resp.StatusCode, resource.URL)
		}

		content := make([]byte, 0)
		buffer := make([]byte, 1024)
		for {
			n, err := resp.Body.Read(buffer)
			if n > 0 {
				content = append(content, buffer[:n]...)
			}
			if err != nil {
				break
			}
		}

		return string(content), nil
	}

	return "", fmt.Errorf("no content source specified for resource %s", resource.URI)
}

// StartStdio starts the MCP server using standard input/output
func (s *MCPServer) StartStdio() error {
	s.logger.Info("Starting MCP server on stdio")
	return server.ServeStdio(s.mcpServer)
}

// Start starts the MCP server on the specified port
func (s *MCPServer) Start(ctx context.Context, port int) error {
	s.logger.WithField("port", port).Info("Starting MCP server")

	// Create HTTP server
	mux := http.NewServeMux()

	// Add JSON-RPC handler for MCP protocol
	jsonrpcHandler := handlers.NewJSONRPCHandler(s.config, s.toolHandler)
	// If OAuth is enabled, wrap with auth and expose discovery
	if s.config.Security.OAuth.Enabled {
		mux.HandleFunc("/.well-known/oauth-protected-resource", s.oauthProtectedResourceHandler(port))
		mux.Handle("/mcp", s.wrapWithAuth(jsonrpcHandler, port))
	} else {
		mux.Handle("/mcp", jsonrpcHandler)
	}

	// Add health check endpoint
	mux.HandleFunc("/health", s.healthCheckHandler)

	// Add metrics endpoint if enabled
	if s.config.Runtime.MetricsEnabled {
		mux.HandleFunc("/metrics", s.metricsHandler)
	}

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	s.logger.WithField("port", port).Info("MCP server started successfully")

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		s.logger.Info("Server context cancelled, shutting down")
		return s.httpServer.Shutdown(context.Background())
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	}
}

// Shutdown gracefully shuts down the server
func (s *MCPServer) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down MCP server")

	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}

	return nil
}

// healthCheckHandler handles health check requests
func (s *MCPServer) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":          "healthy",
		"server_name":     s.config.Server.Name,
		"version":         s.config.Server.Version,
		"tools_count":     len(s.config.Tools),
		"prompts_count":   len(s.config.Prompts),
		"resources_count": len(s.config.Resources),
		"timestamp":       time.Now().UTC().Format(time.RFC3339),
	}

	if err := writeJSON(w, response); err != nil {
		s.logger.WithError(err).Error("Failed to write health check response")
	}
}

// metricsHandler handles metrics requests (basic implementation)
func (s *MCPServer) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	// Basic metrics - in production, you'd use a proper metrics library
	metrics := fmt.Sprintf(`# HELP mcp_server_info Server information
# TYPE mcp_server_info gauge
mcp_server_info{name="%s",version="%s"} 1
# HELP mcp_tools_count Number of registered tools
# TYPE mcp_tools_count gauge
mcp_tools_count %d
# HELP mcp_prompts_count Number of registered prompts  
# TYPE mcp_prompts_count gauge
mcp_prompts_count %d
# HELP mcp_resources_count Number of registered resources
# TYPE mcp_resources_count gauge
mcp_resources_count %d
`,
		s.config.Server.Name,
		s.config.Server.Version,
		len(s.config.Tools),
		len(s.config.Prompts),
		len(s.config.Resources),
	)

	w.Write([]byte(metrics))
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}

// ---- OAuth scaffolding ----

// oauthProtectedResourceHandler serves RFC9728 protected resource metadata so clients can
// discover the authorization servers associated with this MCP server.
func (s *MCPServer) oauthProtectedResourceHandler(port int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		meta := map[string]interface{}{
			"resource":              s.canonicalMCPURL(r, port),
			"authorization_servers": s.config.Security.OAuth.AuthorizationServers,
		}
		_ = json.NewEncoder(w).Encode(meta)
	}
}

// wrapWithAuth validates Authorization: Bearer <token> for /mcp and, when missing or invalid,
// responds with 401 and a WWW-Authenticate header pointing to the protected-resource metadata.
func (s *MCPServer) wrapWithAuth(next http.Handler, port int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz := r.Header.Get("Authorization")
		if authz == "" || !strings.HasPrefix(strings.ToLower(authz), "bearer ") {
			s.writeWWWAuthenticate(w, r, port, "invalid_token", "Missing bearer token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// NOTE: this is a placeholder for a full JWT validation implementation with discovery
		// and JWKS key verification. We surface 401 with proper discovery hints for now.

		// If you add validation: parse token, validate iss/aud/exp using AS metadata & JWKS.
		// On failure, keep the 401 + WWW-Authenticate flow.

		next.ServeHTTP(w, r)
	})
}

func (s *MCPServer) writeWWWAuthenticate(w http.ResponseWriter, r *http.Request, port int, errCode, errDesc string) {
	resourceMeta := s.canonicalBaseURL(r, port) + "/.well-known/oauth-protected-resource"
	val := fmt.Sprintf("Bearer, error=\"%s\", error_description=\"%s\", resource_metadata=\"%s\"", errCode, errDesc, resourceMeta)
	w.Header().Set("WWW-Authenticate", val)
}

func (s *MCPServer) canonicalBaseURL(r *http.Request, port int) string {
	scheme := "https"
	if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") == "" {
		scheme = "http"
	}
	host := r.Host
	if host == "" {
		host = fmt.Sprintf("localhost:%d", port)
	}
	return fmt.Sprintf("%s://%s", scheme, host)
}

func (s *MCPServer) canonicalMCPURL(r *http.Request, port int) string {
	return s.canonicalBaseURL(r, port) + "/mcp"
}
