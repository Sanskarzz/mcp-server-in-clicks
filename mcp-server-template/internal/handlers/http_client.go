package handlers

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"

	"mcp-server-template/internal/config"

	"github.com/sirupsen/logrus"
)

// HTTPClient handles HTTP requests for tool execution
type HTTPClient struct {
	client *http.Client
	logger *logrus.Logger
}

// NewHTTPClient creates a new HTTP client with appropriate configuration
func NewHTTPClient() *HTTPClient {
	// Create HTTP client with reasonable defaults
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:       100,
			IdleConnTimeout:    90 * time.Second,
			DisableCompression: false,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false, // Should be configurable in production
			},
		},
	}

	return &HTTPClient{
		client: client,
		logger: logrus.New(),
	}
}

// ExecuteRequest executes an HTTP request based on tool configuration
func (h *HTTPClient) ExecuteRequest(ctx context.Context, tool *config.ToolConfig, params map[string]interface{}) (*APIResponse, error) {
	// Set timeout for this request
	if tool.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, tool.Timeout.ToDuration())
		defer cancel()
	}
	startTime := time.Now()

	h.logger.WithFields(logrus.Fields{
		"tool_name": tool.Name,
		"endpoint":  tool.Endpoint,
		"method":    tool.Method,
	}).Debug("Executing HTTP request")

	// Execute request with retries
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= tool.Retries; attempt++ {
		// Rebuild request each attempt to avoid issues with consumed bodies
		req, err := h.buildRequest(ctx, tool, params)
		if err != nil {
			return nil, fmt.Errorf("failed to build request: %w", err)
		}
		if attempt > 0 {
			h.logger.WithFields(logrus.Fields{
				"tool_name": tool.Name,
				"attempt":   attempt,
			}).Warn("Retrying request")

			// Exponential backoff
			backoff := time.Duration(attempt) * time.Second
			time.Sleep(backoff)
		}

		resp, lastErr = h.client.Do(req)
		if lastErr == nil && h.isSuccessStatusCode(resp.StatusCode, tool.Validation) {
			break
		}

		if resp != nil {
			resp.Body.Close()
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d attempts: %w", tool.Retries+1, lastErr)
	}

	// Process response
	apiResp, err := h.processResponse(resp, tool)
	if err != nil {
		return nil, fmt.Errorf("failed to process response: %w", err)
	}

	duration := time.Since(startTime)
	h.logger.WithFields(logrus.Fields{
		"tool_name":   tool.Name,
		"status_code": resp.StatusCode,
		"duration_ms": duration.Milliseconds(),
	}).Info("Request completed successfully")

	return apiResp, nil
}

// buildRequest constructs an HTTP request from tool configuration and parameters
func (h *HTTPClient) buildRequest(ctx context.Context, tool *config.ToolConfig, params map[string]interface{}) (*http.Request, error) {
	// Expand endpoint template with params first (e.g., /users/{{.username}})
	expandedEndpoint := tool.Endpoint
	if strings.Contains(expandedEndpoint, "{{") {
		var err error
		expandedEndpoint, err = h.expandTemplate(expandedEndpoint, params)
		if err != nil {
			return nil, fmt.Errorf("failed to expand endpoint template: %w", err)
		}
	}

	// Parse and build URL with query parameters
	parsedURL, err := url.Parse(expandedEndpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint URL: %w", err)
	}

	// Add configured query parameters
	query := parsedURL.Query()
	for key, value := range tool.QueryParams {
		expandedValue, err := h.expandTemplate(value, params)
		if err != nil {
			return nil, fmt.Errorf("failed to expand query param %s: %w", key, err)
		}
		query.Set(key, expandedValue)
	}

	// Add parameter-based query parameters for GET requests
	if strings.ToUpper(tool.Method) == "GET" {
		for _, param := range tool.Parameters {
			if value, exists := params[param.Name]; exists {
				query.Set(param.Name, fmt.Sprintf("%v", value))
			}
		}
	}

	parsedURL.RawQuery = query.Encode()

	// Build request body
	var body io.Reader
	if tool.BodyTemplate != "" && (strings.ToUpper(tool.Method) != "GET") {
		bodyContent, err := h.expandTemplate(tool.BodyTemplate, params)
		if err != nil {
			return nil, fmt.Errorf("failed to expand body template: %w", err)
		}
		body = strings.NewReader(bodyContent)
	} else if strings.ToUpper(tool.Method) != "GET" && len(params) > 0 {
		// Default JSON body for non-GET requests
		jsonBody, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal parameters to JSON: %w", err)
		}
		body = bytes.NewReader(jsonBody)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(tool.Method), parsedURL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type
	if tool.ContentType != "" && body != nil {
		req.Header.Set("Content-Type", tool.ContentType)
	}

	// Set default headers for better API compatibility
	req.Header.Set("User-Agent", "MCP-Server/1.0.0")
	req.Header.Set("Accept", "application/json, text/plain, */*")

	// Add configured headers
	for key, value := range tool.Headers {
		expandedValue, err := h.expandTemplate(value, params)
		if err != nil {
			return nil, fmt.Errorf("failed to expand header %s: %w", key, err)
		}
		req.Header.Set(key, expandedValue)
	}

	// Apply authentication
	if tool.Auth != nil {
		if err := h.applyAuthentication(req, tool.Auth); err != nil {
			return nil, fmt.Errorf("failed to apply authentication: %w", err)
		}
	}

	return req, nil
}

// applyAuthentication applies authentication configuration to the request
func (h *HTTPClient) applyAuthentication(req *http.Request, auth *config.AuthConfig) error {
	switch auth.Type {
	case "bearer":
		token := auth.Token
		if auth.EnvVar != "" {
			if envToken := os.Getenv(auth.EnvVar); envToken != "" {
				token = envToken
			}
		}
		if token == "" {
			return fmt.Errorf("bearer token not found")
		}
		req.Header.Set("Authorization", "Bearer "+token)

	case "basic":
		username := auth.Username
		password := auth.Password
		if auth.EnvVar != "" {
			if envPassword := os.Getenv(auth.EnvVar); envPassword != "" {
				password = envPassword
			}
		}
		if username == "" || password == "" {
			return fmt.Errorf("basic auth credentials not found")
		}
		req.SetBasicAuth(username, password)

	case "api_key":
		for key, value := range auth.Headers {
			finalValue := value
			if auth.EnvVar != "" {
				if envValue := os.Getenv(auth.EnvVar); envValue != "" {
					finalValue = envValue
				}
			}
			req.Header.Set(key, finalValue)
		}

	case "custom":
		for key, value := range auth.Headers {
			req.Header.Set(key, value)
		}
	}

	return nil
}

// expandTemplate expands a template string with parameter values
func (h *HTTPClient) expandTemplate(templateStr string, params map[string]interface{}) (string, error) {
	tmpl, err := template.New("expand").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("invalid template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}

	return buf.String(), nil
}

// processResponse processes the HTTP response and extracts data
func (h *HTTPClient) processResponse(resp *http.Response, tool *config.ToolConfig) (*APIResponse, error) {
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Create API response
	apiResp := &APIResponse{
		StatusCode: resp.StatusCode,
		Headers:    make(map[string]string),
		Body:       string(bodyBytes),
	}

	// Copy response headers
	for key, values := range resp.Header {
		if len(values) > 0 {
			apiResp.Headers[key] = values[0]
		}
	}

	// Parse JSON response if applicable
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") && len(bodyBytes) > 0 {
		var jsonData interface{}
		if err := json.Unmarshal(bodyBytes, &jsonData); err != nil {
			h.logger.WithError(err).Warn("Failed to parse JSON response, returning raw body")
		} else {
			apiResp.Data = jsonData
		}
	}

	// Validate response if validation rules are configured
	if tool.Validation != nil {
		if err := h.validateResponse(apiResp, tool.Validation); err != nil {
			return nil, fmt.Errorf("response validation failed: %w", err)
		}
	}

	return apiResp, nil
}

// isSuccessStatusCode checks if the status code is considered successful
func (h *HTTPClient) isSuccessStatusCode(statusCode int, validation *config.ValidationConfig) bool {
	if validation != nil && len(validation.StatusCodes) > 0 {
		for _, code := range validation.StatusCodes {
			if code == statusCode {
				return true
			}
		}
		return false
	}

	// Default: 2xx status codes are successful
	return statusCode >= 200 && statusCode < 300
}

// validateResponse validates the API response against configured rules
func (h *HTTPClient) validateResponse(resp *APIResponse, validation *config.ValidationConfig) error {
	// Validate status codes
	if len(validation.StatusCodes) > 0 {
		validStatus := false
		for _, code := range validation.StatusCodes {
			if code == resp.StatusCode {
				validStatus = true
				break
			}
		}
		if !validStatus {
			return fmt.Errorf("unexpected status code %d", resp.StatusCode)
		}
	}

	// Validate required fields in JSON response
	if len(validation.RequiredFields) > 0 && resp.Data != nil {
		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			return fmt.Errorf("response is not a JSON object")
		}

		for _, field := range validation.RequiredFields {
			if _, exists := data[field]; !exists {
				return fmt.Errorf("required field %s missing from response", field)
			}
		}
	}

	return nil
}

// APIResponse represents the response from an API call
type APIResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Data       interface{}       `json:"data,omitempty"`
}
