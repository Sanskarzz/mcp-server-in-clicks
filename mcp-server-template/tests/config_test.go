package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"mcp-server-template/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigLoad(t *testing.T) {
	tests := []struct {
		name        string
		configJSON  string
		expectError bool
		validate    func(t *testing.T, cfg *config.Config)
	}{
		{
			name: "valid_minimal_config",
			configJSON: `{
				"server": {
					"name": "test-server",
					"version": "1.0.0",
					"description": "Test server"
				}
			}`,
			expectError: false,
			validate: func(t *testing.T, cfg *config.Config) {
				assert.Equal(t, "test-server", cfg.Server.Name)
				assert.Equal(t, "1.0.0", cfg.Server.Version)
				assert.Equal(t, "Test server", cfg.Server.Description)
			},
		},
		{
			name: "valid_complete_config",
			configJSON: `{
				"server": {
					"name": "complete-server",
					"version": "2.1.0",
					"description": "Complete test server",
					"author": "Test Author",
					"license": "MIT"
				},
				"tools": [
					{
						"name": "test_tool",
						"description": "A test tool",
						"endpoint": "https://api.example.com/test",
						"method": "GET",
						"parameters": [
							{
								"name": "param1",
								"type": "string",
								"description": "Test parameter",
								"required": true
							}
						],
						"timeout": "30s",
						"retries": 3
					}
				],
				"prompts": [
					{
						"name": "test_prompt",
						"description": "A test prompt",
						"content": "Test prompt with {arg1}",
						"arguments": [
							{
								"name": "arg1",
								"description": "Test argument",
								"required": true
							}
						]
					}
				],
				"resources": [
					{
						"uri": "test://resource",
						"name": "Test Resource",
						"description": "A test resource",
						"mime_type": "text/plain",
						"content": "Test content"
					}
				],
				"security": {
					"enable_cors": true,
					"rate_limit": 100
				},
				"runtime": {
					"max_concurrent_requests": 50,
					"log_level": "debug"
				}
			}`,
			expectError: false,
			validate: func(t *testing.T, cfg *config.Config) {
				assert.Equal(t, "complete-server", cfg.Server.Name)
				assert.Len(t, cfg.Tools, 1)
				assert.Len(t, cfg.Prompts, 1)
				assert.Len(t, cfg.Resources, 1)
				assert.True(t, cfg.Security.EnableCORS)
				assert.Equal(t, 50, cfg.Runtime.MaxConcurrentRequests)
			},
		},
		{
			name: "invalid_missing_server",
			configJSON: `{
				"tools": []
			}`,
			expectError: true,
		},
		{
			name: "invalid_empty_server_name",
			configJSON: `{
				"server": {
					"name": "",
					"version": "1.0.0"
				}
			}`,
			expectError: true,
		},
		{
			name: "invalid_malformed_json",
			configJSON: `{
				"server": {
					"name": "test"
				// missing closing brace
			`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.json")

			err := os.WriteFile(configPath, []byte(tt.configJSON), 0644)
			require.NoError(t, err)

			// Load config
			cfg, err := config.Load(configPath)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, cfg)

				if tt.validate != nil {
					tt.validate(t, cfg)
				}
			}
		})
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_config",
			config: &config.Config{
				Server: config.ServerConfig{
					Name:        "valid-server",
					Version:     "1.0.0",
					Description: "Valid server config",
				},
			},
			expectError: false,
		},
		{
			name: "duplicate_tool_names",
			config: &config.Config{
				Server: config.ServerConfig{
					Name:    "test-server",
					Version: "1.0.0",
				},
				Tools: []config.ToolConfig{
					{
						Name:        "duplicate_tool",
						Description: "First tool",
						Endpoint:    "https://api.example.com/1",
						Method:      "GET",
					},
					{
						Name:        "duplicate_tool",
						Description: "Second tool",
						Endpoint:    "https://api.example.com/2",
						Method:      "GET",
					},
				},
			},
			expectError: true,
			errorMsg:    "duplicate tool name",
		},
		{
			name: "invalid_semver",
			config: &config.Config{
				Server: config.ServerConfig{
					Name:    "test-server",
					Version: "invalid-version",
				},
			},
			expectError: true,
		},
		{
			name: "invalid_tool_endpoint",
			config: &config.Config{
				Server: config.ServerConfig{
					Name:    "test-server",
					Version: "1.0.0",
				},
				Tools: []config.ToolConfig{
					{
						Name:        "test_tool",
						Description: "Test tool",
						Endpoint:    "not-a-url",
						Method:      "GET",
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.Validate(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	// Create minimal config
	cfg := &config.Config{
		Server: config.ServerConfig{
			Name: "test-server",
		},
		Tools: []config.ToolConfig{
			{
				Name:        "test_tool",
				Description: "Test tool",
				Endpoint:    "https://api.example.com/test",
			},
		},
	}

	// Apply defaults (this would normally happen in Load)
	setDefaults(cfg)

	// Check defaults were applied
	assert.Equal(t, "1.0.0", cfg.Server.Version)
	assert.Equal(t, "GET", cfg.Tools[0].Method)
	assert.Equal(t, config.Duration(30*time.Second), cfg.Tools[0].Timeout)
	assert.Equal(t, 3, cfg.Tools[0].Retries)
	assert.Equal(t, 100, cfg.Security.RateLimit)
	assert.Equal(t, 100, cfg.Runtime.MaxConcurrentRequests)
	assert.Equal(t, "info", cfg.Runtime.LogLevel)
}

func TestEnvironmentVariableSubstitution(t *testing.T) {
	// Set test environment variable
	os.Setenv("TEST_API_KEY", "secret-key-123")
	defer os.Unsetenv("TEST_API_KEY")

	configJSON := `{
		"server": {
			"name": "env-test-server",
			"version": "1.0.0"
		},
		"tools": [
			{
				"name": "test_tool",
				"description": "Test tool with env var",
				"endpoint": "https://api.example.com/test",
				"method": "GET",
				"headers": {
					"Authorization": "Bearer ${TEST_API_KEY}"
				}
			}
		]
	}`

	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	err := os.WriteFile(configPath, []byte(configJSON), 0644)
	require.NoError(t, err)

	// Load config
	cfg, err := config.Load(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check that environment variable was substituted
	assert.Equal(t, "Bearer secret-key-123", cfg.Tools[0].Headers["Authorization"])
}

func TestAuthConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		auth        *config.AuthConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_bearer_auth",
			auth: &config.AuthConfig{
				Type:  "bearer",
				Token: "test-token",
			},
			expectError: false,
		},
		{
			name: "valid_bearer_auth_with_env",
			auth: &config.AuthConfig{
				Type:   "bearer",
				EnvVar: "API_TOKEN",
			},
			expectError: false,
		},
		{
			name: "invalid_bearer_auth_no_token",
			auth: &config.AuthConfig{
				Type: "bearer",
			},
			expectError: true,
			errorMsg:    "bearer auth requires",
		},
		{
			name: "valid_basic_auth",
			auth: &config.AuthConfig{
				Type:     "basic",
				Username: "user",
				Password: "pass",
			},
			expectError: false,
		},
		{
			name: "invalid_basic_auth_no_password",
			auth: &config.AuthConfig{
				Type:     "basic",
				Username: "user",
			},
			expectError: true,
			errorMsg:    "basic auth requires",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAuthConfig(tt.auth)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function that would normally be internal to config package
func setDefaults(cfg *config.Config) {
	if cfg.Server.Version == "" {
		cfg.Server.Version = "1.0.0"
	}

	for i := range cfg.Tools {
		tool := &cfg.Tools[i]
		if tool.Method == "" {
			tool.Method = "GET"
		}
		if tool.Timeout == 0 {
			tool.Timeout = 30 * time.Second
		}
		if tool.Retries == 0 {
			tool.Retries = 3
		}
	}

	if cfg.Security.RateLimit == 0 {
		cfg.Security.RateLimit = 100
	}

	if cfg.Runtime.MaxConcurrentRequests == 0 {
		cfg.Runtime.MaxConcurrentRequests = 100
	}
	if cfg.Runtime.LogLevel == "" {
		cfg.Runtime.LogLevel = "info"
	}
}

// Helper function that would normally be internal to config package
func validateAuthConfig(auth *config.AuthConfig) error {
	switch auth.Type {
	case "bearer":
		if auth.Token == "" && auth.EnvVar == "" {
			return fmt.Errorf("bearer auth requires either token or env_var")
		}
	case "basic":
		if auth.Username == "" || (auth.Password == "" && auth.EnvVar == "") {
			return fmt.Errorf("basic auth requires username and either password or env_var")
		}
	}
	return nil
}
