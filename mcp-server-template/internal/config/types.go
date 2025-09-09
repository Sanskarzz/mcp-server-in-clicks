package config

import (
	"encoding/json"
	"fmt"
	"time"
)

// Config represents the complete configuration for an MCP server instance
type Config struct {
	Server    ServerConfig     `json:"server" validate:"required"`
	Tools     []ToolConfig     `json:"tools"`
	Prompts   []PromptConfig   `json:"prompts"`
	Resources []ResourceConfig `json:"resources"`
	Security  SecurityConfig   `json:"security"`
	Runtime   RuntimeConfig    `json:"runtime"`
}

// ServerConfig defines the basic server metadata and configuration
type ServerConfig struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Version     string `json:"version" validate:"required,semver"`
	Description string `json:"description" validate:"max=500"`
	Author      string `json:"author" validate:"max=100"`
	License     string `json:"license" validate:"max=50"`
}

// ToolConfig defines a single tool that makes HTTP API calls
type ToolConfig struct {
	Name          string            `json:"name" validate:"required,min=1,max=100"`
	Description   string            `json:"description" validate:"required,min=1,max=500"`
	Endpoint      string            `json:"endpoint" validate:"required,url"`
	Method        string            `json:"method" validate:"required,oneof=GET POST PUT PATCH DELETE HEAD OPTIONS"`
	Headers       map[string]string `json:"headers"`
	QueryParams   map[string]string `json:"query_params"`
	BodyTemplate  string            `json:"body_template"`
	ContentType   string            `json:"content_type" validate:"omitempty,oneof=application/json application/xml text/plain application/x-www-form-urlencoded"`
	Parameters    []ParameterConfig `json:"parameters"`
	ReturnType    string            `json:"return_type" validate:"omitempty,oneof=string number boolean object array"`
	Timeout       Duration          `json:"timeout"`
	Retries       int               `json:"retries" validate:"min=0,max=5"`
	Auth          *AuthConfig       `json:"auth,omitempty"`
	Validation    *ValidationConfig `json:"validation,omitempty"`
	UpstreamOAuth *OAuth2Config     `json:"upstream_oauth,omitempty"`
}

// ParameterConfig defines input parameters for tools
type ParameterConfig struct {
	Name        string               `json:"name" validate:"required,min=1,max=50"`
	Type        string               `json:"type" validate:"required,oneof=string number boolean object array"`
	Description string               `json:"description" validate:"required,min=1,max=200"`
	Required    bool                 `json:"required"`
	Default     interface{}          `json:"default"`
	Validation  *ParameterValidation `json:"validation,omitempty"`
}

// ParameterValidation defines validation rules for parameters
type ParameterValidation struct {
	MinLength *int     `json:"min_length,omitempty"`
	MaxLength *int     `json:"max_length,omitempty"`
	Pattern   *string  `json:"pattern,omitempty"`
	MinValue  *float64 `json:"min_value,omitempty"`
	MaxValue  *float64 `json:"max_value,omitempty"`
	Enum      []string `json:"enum,omitempty"`
}

// AuthConfig defines authentication settings for API calls
type AuthConfig struct {
	Type     string            `json:"type" validate:"required,oneof=bearer basic api_key custom"`
	Token    string            `json:"token,omitempty"`
	Username string            `json:"username,omitempty"`
	Password string            `json:"password,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
	EnvVar   string            `json:"env_var,omitempty"` // Environment variable name for token
}

// OAuth2Config describes how to acquire an upstream access token to call a tool endpoint
type OAuth2Config struct {
	GrantType       string   `json:"grant_type"` // currently supports "client_credentials"
	Issuer          string   `json:"issuer,omitempty"`
	TokenURL        string   `json:"token_url,omitempty"`
	ClientID        string   `json:"client_id,omitempty"`
	ClientSecret    string   `json:"client_secret,omitempty"`
	ClientIDEnv     string   `json:"client_id_env,omitempty"`
	ClientSecretEnv string   `json:"client_secret_env,omitempty"`
	Scopes          []string `json:"scopes,omitempty"`
	Audience        string   `json:"audience,omitempty"`
	CacheTTL        Duration `json:"cache_ttl,omitempty"`
}

// ValidationConfig defines response validation rules
type ValidationConfig struct {
	Schema         string   `json:"schema,omitempty"`          // JSON schema for response validation
	StatusCodes    []int    `json:"status_codes,omitempty"`    // Expected HTTP status codes
	RequiredFields []string `json:"required_fields,omitempty"` // Required fields in response
}

// PromptConfig defines static prompts for the MCP server
type PromptConfig struct {
	Name        string           `json:"name" validate:"required,min=1,max=100"`
	Description string           `json:"description" validate:"required,min=1,max=500"`
	Content     string           `json:"content" validate:"required,min=1"`
	Arguments   []ArgumentConfig `json:"arguments"`
}

// ArgumentConfig defines prompt arguments
type ArgumentConfig struct {
	Name        string `json:"name" validate:"required,min=1,max=50"`
	Description string `json:"description" validate:"required,min=1,max=200"`
	Required    bool   `json:"required"`
}

// ResourceConfig defines static resources served by the MCP server
type ResourceConfig struct {
	URI         string `json:"uri" validate:"required,uri"`
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
	MimeType    string `json:"mime_type" validate:"required"`
	Content     string `json:"content,omitempty"`   // Inline content
	FilePath    string `json:"file_path,omitempty"` // Path to file
	URL         string `json:"url,omitempty"`       // External URL
}

// SecurityConfig defines security settings for the server
type SecurityConfig struct {
	EnableCORS      bool        `json:"enable_cors"`
	AllowedOrigins  []string    `json:"allowed_origins"`
	EnableRateLimit bool        `json:"enable_rate_limit"`
	RateLimit       int         `json:"rate_limit" validate:"min=1,max=10000"`
	EnableAuth      bool        `json:"enable_auth"`
	APIKeys         []string    `json:"api_keys"`
	TLSCertPath     string      `json:"tls_cert_path"`
	TLSKeyPath      string      `json:"tls_key_path"`
	OAuth           OAuthConfig `json:"oauth"`
}

// OAuthConfig configures OAuth/OIDC-based authorization for the MCP HTTP transport
type OAuthConfig struct {
	// Enable transport-level OAuth for /mcp endpoint
	Enabled bool `json:"enabled"`
	// One or more issuer base URLs (what the user provides in your UI)
	AuthorizationServers []string `json:"authorization_servers"`
	// Accept tokens whose audience matches one of these values. When empty, we compute
	// the canonical MCP endpoint URL at request-time and validate against that.
	AcceptedAudiences []string `json:"accepted_audiences"`
	// Optional scopes your server requires to access /mcp
	RequiredScopes []string `json:"required_scopes"`
	// JWKS cache TTL for key rotation
	JWKSCacheTTL Duration `json:"jwks_cache_ttl"`
	// Development only: allow HTTP discovery (not recommended in prod)
	AllowInsecureHTTP bool `json:"allow_insecure_http"`
}

// RuntimeConfig defines runtime behavior settings
type RuntimeConfig struct {
	MaxConcurrentRequests int      `json:"max_concurrent_requests" validate:"min=1,max=1000"`
	DefaultTimeout        Duration `json:"default_timeout"`
	HealthCheckInterval   Duration `json:"health_check_interval"`
	MetricsEnabled        bool     `json:"metrics_enabled"`
	LogLevel              string   `json:"log_level" validate:"oneof=debug info warn error"`
	Environment           string   `json:"environment" validate:"oneof=development staging production"`
}

// Duration is a wrapper around time.Duration for JSON marshaling
type Duration time.Duration

// UnmarshalJSON implements the json.Unmarshaler interface
func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		*d = Duration(0)
		return nil
	}
	duration, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration format: %w", err)
	}
	*d = Duration(duration)
	return nil
}

// MarshalJSON implements the json.Marshaler interface
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// String returns the duration as a string
func (d Duration) String() string {
	return time.Duration(d).String()
}

// ToDuration converts to time.Duration
func (d Duration) ToDuration() time.Duration {
	return time.Duration(d)
}
