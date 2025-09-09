package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"mcp-server-template/internal/config"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"
)

// JSONRPCHandler handles MCP JSON-RPC requests over HTTP
type JSONRPCHandler struct {
	config      *config.Config
	toolHandler *ToolHandler
	logger      *logrus.Logger
	mcpServer   interface{} // Store reference to MCP server if needed
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewJSONRPCHandler creates a new JSON-RPC handler
func NewJSONRPCHandler(cfg *config.Config, toolHandler *ToolHandler) *JSONRPCHandler {
	return &JSONRPCHandler{
		config:      cfg,
		toolHandler: toolHandler,
		logger:      logrus.New(),
	}
}

// ServeHTTP implements http.Handler for JSON-RPC requests
func (h *JSONRPCHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Always set CORS headers for web clients like Cursor
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Max-Age", "86400")

	// Handle preflight OPTIONS request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPost {
		h.writeError(w, nil, -32600, "Invalid Request", "Only POST method is allowed")
		return
	}

	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, nil, -32700, "Parse error", err.Error())
		return
	}

	h.logger.WithFields(logrus.Fields{
		"method": req.Method,
		"id":     req.ID,
	}).Debug("Handling JSON-RPC request")

	// Handle different MCP methods
	switch req.Method {
	case "initialize":
		h.handleInitialize(w, &req)
	case "initialized":
		h.handleInitialized(w, &req)
	case "tools/list":
		h.handleToolsList(w, &req)
	case "tools/call":
		h.handleToolsCall(w, &req)
	case "prompts/list":
		h.handlePromptsList(w, &req)
	case "prompts/get":
		h.handlePromptsGet(w, &req)
	case "resources/list":
		h.handleResourcesList(w, &req)
	case "resources/read":
		h.handleResourcesRead(w, &req)
	case "ping":
		h.handlePing(w, &req)
	default:
		h.writeError(w, req.ID, -32601, "Method not found", fmt.Sprintf("Unknown method: %s", req.Method))
	}
}

func (h *JSONRPCHandler) handleInitialize(w http.ResponseWriter, req *JSONRPCRequest) {
	// Parse initialize params
	var params struct {
		ProtocolVersion string `json:"protocolVersion"`
		Capabilities    struct {
			Tools     map[string]interface{} `json:"tools,omitempty"`
			Prompts   map[string]interface{} `json:"prompts,omitempty"`
			Resources map[string]interface{} `json:"resources,omitempty"`
		} `json:"capabilities"`
		ClientInfo struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"clientInfo"`
	}

	if req.Params != nil {
		paramBytes, _ := json.Marshal(req.Params)
		json.Unmarshal(paramBytes, &params)
	}

	h.logger.WithFields(logrus.Fields{
		"client_name":      params.ClientInfo.Name,
		"client_version":   params.ClientInfo.Version,
		"protocol_version": params.ProtocolVersion,
	}).Info("MCP client initializing")

	// Return server capabilities
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": true,
			},
			"prompts": map[string]interface{}{
				"listChanged": true,
			},
			"resources": map[string]interface{}{
				"listChanged": true,
			},
		},
		"serverInfo": map[string]interface{}{
			"name":    h.config.Server.Name,
			"version": h.config.Server.Version,
		},
		"instructions": "MCP Server ready for tool, prompt, and resource operations",
	}

	h.writeSuccess(w, req.ID, result)
}

func (h *JSONRPCHandler) handleInitialized(w http.ResponseWriter, req *JSONRPCRequest) {
	h.logger.Info("MCP client initialized")
	// Return empty success response for initialized notification
	h.writeSuccess(w, req.ID, map[string]interface{}{})
}

func (h *JSONRPCHandler) handleToolsList(w http.ResponseWriter, req *JSONRPCRequest) {
	h.logger.Debug("Listing available tools")

	tools := make([]map[string]interface{}, 0, len(h.config.Tools))
	for _, tool := range h.config.Tools {
		// Build input schema
		properties := make(map[string]interface{})
		required := make([]string, 0)

		for _, param := range tool.Parameters {
			propSchema := map[string]interface{}{
				"type":        param.Type,
				"description": param.Description,
			}

			if param.Default != nil {
				propSchema["default"] = param.Default
			}

			// Add validation constraints
			if param.Validation != nil {
				if param.Type == "string" {
					if param.Validation.MinLength != nil {
						propSchema["minLength"] = *param.Validation.MinLength
					}
					if param.Validation.MaxLength != nil {
						propSchema["maxLength"] = *param.Validation.MaxLength
					}
					if param.Validation.Pattern != nil {
						propSchema["pattern"] = *param.Validation.Pattern
					}
					if len(param.Validation.Enum) > 0 {
						propSchema["enum"] = param.Validation.Enum
					}
				} else if param.Type == "number" {
					if param.Validation.MinValue != nil {
						propSchema["minimum"] = *param.Validation.MinValue
					}
					if param.Validation.MaxValue != nil {
						propSchema["maximum"] = *param.Validation.MaxValue
					}
				}
			}

			properties[param.Name] = propSchema
			if param.Required {
				required = append(required, param.Name)
			}
		}

		inputSchema := map[string]interface{}{
			"type":       "object",
			"properties": properties,
		}
		if len(required) > 0 {
			inputSchema["required"] = required
		}

		toolDef := map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": inputSchema,
		}

		tools = append(tools, toolDef)
	}

	result := map[string]interface{}{
		"tools": tools,
	}

	h.writeSuccess(w, req.ID, result)
}

func (h *JSONRPCHandler) handleToolsCall(w http.ResponseWriter, req *JSONRPCRequest) {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if req.Params != nil {
		paramBytes, _ := json.Marshal(req.Params)
		if err := json.Unmarshal(paramBytes, &params); err != nil {
			h.writeError(w, req.ID, -32602, "Invalid params", err.Error())
			return
		}
	}

	h.logger.WithFields(logrus.Fields{
		"tool_name": params.Name,
		"arguments": params.Arguments,
	}).Info("Executing tool")

	// Execute the tool using our tool handler with shorter timeout for testing
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := h.toolHandler.ExecuteTool(ctx, params.Name, params.Arguments)
	if err != nil {
		h.logger.WithError(err).WithField("tool_name", params.Name).Error("Tool execution failed")
		// Return a more user-friendly error for testing
		h.writeError(w, req.ID, -32000, "Tool execution error", fmt.Sprintf("Failed to execute tool '%s': %s", params.Name, err.Error()))
		return
	}

	// Convert mcp.CallToolResult to JSON-RPC format
	if result.IsError {
		// Extract error message from content
		errorMsg := "Tool execution failed"
		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
				errorMsg = textContent.Text
			} else if textVal, ok := result.Content[0].(mcp.TextContent); ok {
				errorMsg = textVal.Text
			} else if m, ok := result.Content[0].(map[string]interface{}); ok {
				if t, ok := m["text"].(string); ok {
					errorMsg = t
				}
			}
		}
		h.writeError(w, req.ID, -32000, "Tool execution error", errorMsg)
		return
	}

	// Convert successful result
	h.logger.WithField("content_len", len(result.Content)).Debug("Converting tool result content")
	// Be lenient about content element types. Different SDK versions may use
	// pointer or value receivers, or even maps for content. We normalize to
	// JSON-RPC text content objects.
	content := make([]map[string]interface{}, 0, len(result.Content))
	for _, c := range result.Content {
		h.logger.WithField("elem_type", fmt.Sprintf("%T", c)).Debug("Result content element type")
		// 1) Pointer form
		if textPtr, ok := c.(*mcp.TextContent); ok {
			content = append(content, map[string]interface{}{
				"type": "text",
				"text": textPtr.Text,
			})
			continue
		}

		// 2) Value form
		if textVal, ok := c.(mcp.TextContent); ok {
			content = append(content, map[string]interface{}{
				"type": "text",
				"text": textVal.Text,
			})
			continue
		}

		// 3) Map form {type:"text", text:"..."}
		if m, ok := c.(map[string]interface{}); ok {
			if m["type"] == "text" {
				if t, ok := m["text"].(string); ok {
					content = append(content, map[string]interface{}{
						"type": "text",
						"text": t,
					})
					continue
				}
			}
		}

		// 4) Fallback: stringify unknown content kinds
		content = append(content, map[string]interface{}{
			"type": "text",
			"text": fmt.Sprintf("%v", c),
		})
	}

	response := map[string]interface{}{
		"content": content,
	}

	h.writeSuccess(w, req.ID, response)
}

func (h *JSONRPCHandler) handlePromptsList(w http.ResponseWriter, req *JSONRPCRequest) {
	h.logger.Debug("Listing available prompts")

	prompts := make([]map[string]interface{}, 0, len(h.config.Prompts))
	for _, prompt := range h.config.Prompts {
		arguments := make([]map[string]interface{}, 0, len(prompt.Arguments))
		for _, arg := range prompt.Arguments {
			arguments = append(arguments, map[string]interface{}{
				"name":        arg.Name,
				"description": arg.Description,
				"required":    arg.Required,
			})
		}

		promptDef := map[string]interface{}{
			"name":        prompt.Name,
			"description": prompt.Description,
			"arguments":   arguments,
		}

		prompts = append(prompts, promptDef)
	}

	result := map[string]interface{}{
		"prompts": prompts,
	}

	h.writeSuccess(w, req.ID, result)
}

func (h *JSONRPCHandler) handlePromptsGet(w http.ResponseWriter, req *JSONRPCRequest) {
	var params struct {
		Name      string            `json:"name"`
		Arguments map[string]string `json:"arguments"`
	}

	if req.Params != nil {
		paramBytes, _ := json.Marshal(req.Params)
		json.Unmarshal(paramBytes, &params)
	}

	h.logger.WithFields(logrus.Fields{
		"prompt_name": params.Name,
		"arguments":   params.Arguments,
	}).Info("Getting prompt")

	// Find the prompt
	var promptConfig *config.PromptConfig
	for _, p := range h.config.Prompts {
		if p.Name == params.Name {
			promptConfig = &p
			break
		}
	}

	if promptConfig == nil {
		h.writeError(w, req.ID, -32602, "Invalid params", fmt.Sprintf("Prompt '%s' not found", params.Name))
		return
	}

	// Substitute arguments in the prompt content
	content := promptConfig.Content
	for key, value := range params.Arguments {
		placeholder := fmt.Sprintf("{%s}", key)
		content = strings.ReplaceAll(content, placeholder, value)
	}

	result := map[string]interface{}{
		"description": promptConfig.Description,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": map[string]interface{}{
					"type": "text",
					"text": content,
				},
			},
		},
	}

	h.writeSuccess(w, req.ID, result)
}

func (h *JSONRPCHandler) handleResourcesList(w http.ResponseWriter, req *JSONRPCRequest) {
	h.logger.Debug("Listing available resources")

	resources := make([]map[string]interface{}, 0, len(h.config.Resources))
	for _, resource := range h.config.Resources {
		resourceDef := map[string]interface{}{
			"uri":         resource.URI,
			"name":        resource.Name,
			"description": resource.Description,
			"mimeType":    resource.MimeType,
		}

		resources = append(resources, resourceDef)
	}

	result := map[string]interface{}{
		"resources": resources,
	}

	h.writeSuccess(w, req.ID, result)
}

func (h *JSONRPCHandler) handleResourcesRead(w http.ResponseWriter, req *JSONRPCRequest) {
	var params struct {
		URI string `json:"uri"`
	}

	if req.Params != nil {
		paramBytes, _ := json.Marshal(req.Params)
		json.Unmarshal(paramBytes, &params)
	}

	h.logger.WithField("uri", params.URI).Info("Reading resource")

	// Find the resource
	var resourceConfig *config.ResourceConfig
	for _, r := range h.config.Resources {
		if r.URI == params.URI {
			resourceConfig = &r
			break
		}
	}

	if resourceConfig == nil {
		h.writeError(w, req.ID, -32602, "Invalid params", fmt.Sprintf("Resource '%s' not found", params.URI))
		return
	}

	// Get resource content
	content := resourceConfig.Content
	if content == "" && resourceConfig.FilePath != "" {
		// Could read from file here if needed
		content = "File content would be loaded here"
	}
	if content == "" && resourceConfig.URL != "" {
		// Could fetch from URL here if needed
		content = "URL content would be fetched here"
	}

	result := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"uri":      resourceConfig.URI,
				"mimeType": resourceConfig.MimeType,
				"text":     content,
			},
		},
	}

	h.writeSuccess(w, req.ID, result)
}

func (h *JSONRPCHandler) handlePing(w http.ResponseWriter, req *JSONRPCRequest) {
	h.writeSuccess(w, req.ID, map[string]interface{}{})
}

func (h *JSONRPCHandler) writeSuccess(w http.ResponseWriter, id interface{}, result interface{}) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *JSONRPCHandler) writeError(w http.ResponseWriter, id interface{}, code int, message string, data interface{}) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // JSON-RPC errors still use 200 OK
	json.NewEncoder(w).Encode(response)
}
