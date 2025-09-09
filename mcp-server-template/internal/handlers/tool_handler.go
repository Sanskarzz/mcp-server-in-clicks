package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"mcp-server-template/internal/config"
	"mcp-server-template/internal/validation"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
)

// ToolHandler manages dynamic tool registration and execution
type ToolHandler struct {
	httpClient *HTTPClient
	validator  *validation.Validator
	logger     *logrus.Logger
	tools      map[string]*config.ToolConfig
}

// NewToolHandler creates a new tool handler
func NewToolHandler() *ToolHandler {
	return &ToolHandler{
		httpClient: NewHTTPClient(),
		validator:  validation.New(),
		logger:     logrus.New(),
		tools:      make(map[string]*config.ToolConfig),
	}
}

// RegisterTools registers all configured tools with the MCP server
func (h *ToolHandler) RegisterTools(mcpServer *server.MCPServer, tools []config.ToolConfig) error {
	h.logger.WithField("tools_count", len(tools)).Info("Registering tools")

	for _, tool := range tools {
		// Store tool configuration for later use
		h.tools[tool.Name] = &tool

		// Create the MCP tool using the builder pattern
		var toolOpts []mcp.ToolOption
		toolOpts = append(toolOpts, mcp.WithDescription(tool.Description))

		// Add parameters using the tool options
		for _, param := range tool.Parameters {
			switch param.Type {
			case "string":
				var opts []mcp.PropertyOption
				opts = append(opts, mcp.Description(param.Description))
				if param.Required {
					opts = append(opts, mcp.Required())
				}
				if param.Validation != nil {
					if param.Validation.MinLength != nil {
						opts = append(opts, mcp.MinLength(*param.Validation.MinLength))
					}
					if param.Validation.MaxLength != nil {
						opts = append(opts, mcp.MaxLength(*param.Validation.MaxLength))
					}
					if param.Validation.Pattern != nil {
						opts = append(opts, mcp.Pattern(*param.Validation.Pattern))
					}
					if len(param.Validation.Enum) > 0 {
						opts = append(opts, mcp.Enum(param.Validation.Enum...))
					}
				}
				toolOpts = append(toolOpts, mcp.WithString(param.Name, opts...))
			case "number":
				var opts []mcp.PropertyOption
				opts = append(opts, mcp.Description(param.Description))
				if param.Required {
					opts = append(opts, mcp.Required())
				}
				if param.Validation != nil {
					if param.Validation.MinValue != nil {
						opts = append(opts, mcp.Min(*param.Validation.MinValue))
					}
					if param.Validation.MaxValue != nil {
						opts = append(opts, mcp.Max(*param.Validation.MaxValue))
					}
				}
				toolOpts = append(toolOpts, mcp.WithNumber(param.Name, opts...))
			case "boolean":
				var opts []mcp.PropertyOption
				opts = append(opts, mcp.Description(param.Description))
				if param.Required {
					opts = append(opts, mcp.Required())
				}
				toolOpts = append(toolOpts, mcp.WithBoolean(param.Name, opts...))
			}
		}

		mcpTool := mcp.NewTool(tool.Name, toolOpts...)

		// Register the tool with the MCP server using the modern API
		mcpServer.AddTool(mcpTool, func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			return h.ExecuteTool(context.Background(), tool.Name, arguments)
		})

		h.logger.WithFields(logrus.Fields{
			"tool_name": tool.Name,
			"endpoint":  tool.Endpoint,
			"method":    tool.Method,
		}).Debug("Tool registered successfully")
	}

	h.logger.Info("All tools registered successfully")
	return nil
}

// ExecuteTool executes a tool with the given parameters
func (h *ToolHandler) ExecuteTool(ctx context.Context, toolName string, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	h.logger.WithFields(logrus.Fields{
		"tool_name": toolName,
		"arguments": h.sanitizeArguments(arguments),
	}).Info("Executing tool")

	// Get tool configuration
	tool, exists := h.tools[toolName]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", toolName)
	}

	// Validate input parameters
	if err := h.validateParameters(tool, arguments); err != nil {
		return nil, fmt.Errorf("parameter validation failed: %w", err)
	}

	// Execute the HTTP request
	response, err := h.httpClient.ExecuteRequest(ctx, tool, arguments)
	if err != nil {
		h.logger.WithError(err).WithField("tool_name", toolName).Error("Tool execution failed")
		// Return precise, actionable error text for LLMs/clients
		return mcp.NewToolResultError(fmt.Sprintf("%s %s failed: %s", tool.Method, tool.Endpoint, err.Error())), nil
	}

	// Convert response to MCP result
	result := h.convertResponseToMCPResult(response, tool)

	h.logger.WithFields(logrus.Fields{
		"tool_name":   toolName,
		"status_code": response.StatusCode,
	}).Info("Tool executed successfully")

	return result, nil
}

// validateParameters validates input parameters against tool configuration
func (h *ToolHandler) validateParameters(tool *config.ToolConfig, arguments map[string]interface{}) error {
	// Check required parameters
	for _, param := range tool.Parameters {
		value, exists := arguments[param.Name]

		if param.Required && !exists {
			return fmt.Errorf("required parameter %s is missing", param.Name)
		}

		if exists {
			// Validate parameter type and constraints
			if err := h.validateParameterValue(&param, value); err != nil {
				return fmt.Errorf("parameter %s validation failed: %w", param.Name, err)
			}
		} else if param.Default != nil {
			// Use default value if parameter is not provided
			arguments[param.Name] = param.Default
		}
	}

	return nil
}

// validateParameterValue validates a single parameter value
func (h *ToolHandler) validateParameterValue(param *config.ParameterConfig, value interface{}) error {
	// Type validation
	switch param.Type {
	case "string":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", value)
		}

		if param.Validation != nil {
			if param.Validation.MinLength != nil && len(str) < *param.Validation.MinLength {
				return fmt.Errorf("string too short, minimum length is %d", *param.Validation.MinLength)
			}
			if param.Validation.MaxLength != nil && len(str) > *param.Validation.MaxLength {
				return fmt.Errorf("string too long, maximum length is %d", *param.Validation.MaxLength)
			}
			if param.Validation.Pattern != nil {
				matched, err := regexp.MatchString(*param.Validation.Pattern, str)
				if err != nil {
					return fmt.Errorf("invalid pattern: %w", err)
				}
				if !matched {
					return fmt.Errorf("string does not match pattern %s", *param.Validation.Pattern)
				}
			}
			if len(param.Validation.Enum) > 0 {
				validValue := false
				for _, enumValue := range param.Validation.Enum {
					if str == enumValue {
						validValue = true
						break
					}
				}
				if !validValue {
					return fmt.Errorf("value must be one of: %v", param.Validation.Enum)
				}
			}
		}

	case "number":
		var num float64
		switch v := value.(type) {
		case float64:
			num = v
		case int:
			num = float64(v)
		case string:
			var err error
			num, err = strconv.ParseFloat(v, 64)
			if err != nil {
				return fmt.Errorf("cannot convert string to number: %w", err)
			}
		default:
			return fmt.Errorf("expected number, got %T", value)
		}

		if param.Validation != nil {
			if param.Validation.MinValue != nil && num < *param.Validation.MinValue {
				return fmt.Errorf("number too small, minimum value is %f", *param.Validation.MinValue)
			}
			if param.Validation.MaxValue != nil && num > *param.Validation.MaxValue {
				return fmt.Errorf("number too large, maximum value is %f", *param.Validation.MaxValue)
			}
		}

	case "boolean":
		_, ok := value.(bool)
		if !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}

	case "object":
		_, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected object, got %T", value)
		}

	case "array":
		_, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
	}

	return nil
}

// convertResponseToMCPResult converts an API response to MCP result format
func (h *ToolHandler) convertResponseToMCPResult(response *APIResponse, tool *config.ToolConfig) *mcp.CallToolResult {
	// Determine if the response indicates an error
	if response.StatusCode >= 400 {
		return mcp.NewToolResultError(fmt.Sprintf("HTTP Error %d: %s", response.StatusCode, response.Body))
	}

	// Format response based on tool configuration
	switch tool.ReturnType {
	case "string":
		return mcp.NewToolResultText(response.Body)

	case "object", "array":
		if response.Data != nil {
			// Return structured data as JSON
			jsonBytes, err := json.MarshalIndent(response.Data, "", "  ")
			if err != nil {
				return mcp.NewToolResultText(response.Body)
			} else {
				return mcp.NewToolResultText(string(jsonBytes))
			}
		} else {
			return mcp.NewToolResultText(response.Body)
		}

	default:
		// Default: return the response body as text
		if response.Data != nil {
			jsonBytes, err := json.MarshalIndent(response.Data, "", "  ")
			if err != nil {
				return mcp.NewToolResultText(response.Body)
			} else {
				return mcp.NewToolResultText(string(jsonBytes))
			}
		} else {
			return mcp.NewToolResultText(response.Body)
		}
	}
}

// sanitizeArguments removes sensitive data from arguments for logging
func (h *ToolHandler) sanitizeArguments(arguments map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})

	sensitiveKeys := []string{"password", "token", "api_key", "secret", "auth"}

	for key, value := range arguments {
		// Check if the key contains sensitive information
		isSensitive := false
		for _, sensitiveKey := range sensitiveKeys {
			if regexp.MustCompile(`(?i)` + sensitiveKey).MatchString(key) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			sanitized[key] = "***REDACTED***"
		} else {
			sanitized[key] = value
		}
	}

	return sanitized
}
