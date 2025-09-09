package config

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom validators
	validate.RegisterValidation("semver", validateSemVer)
}

// Load reads and parses a configuration file
func Load(configPath string) (*Config, error) {
	logrus.WithField("config_path", configPath).Debug("Loading configuration")

	// Read configuration file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Perform environment variable substitution
	configContent := substituteEnvVars(string(data))

	// Parse JSON configuration
	var cfg Config
	if err := json.Unmarshal([]byte(configContent), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Set default values
	setDefaults(&cfg)

	logrus.WithFields(logrus.Fields{
		"server_name":     cfg.Server.Name,
		"tools_count":     len(cfg.Tools),
		"prompts_count":   len(cfg.Prompts),
		"resources_count": len(cfg.Resources),
	}).Info("Configuration loaded successfully")

	return &cfg, nil
}

// Validate validates the configuration using struct tags and business logic
func Validate(cfg *Config) error {
	logrus.Debug("Validating configuration")

	// Struct validation using tags
	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Business logic validation
	if err := validateBusinessRules(cfg); err != nil {
		return fmt.Errorf("business rule validation failed: %w", err)
	}

	logrus.Debug("Configuration validation passed")
	return nil
}

// substituteEnvVars replaces ${VAR_NAME} patterns with environment variable values
func substituteEnvVars(content string) string {
	envVarRegex := regexp.MustCompile(`\${([^}]+)}`)

	return envVarRegex.ReplaceAllStringFunc(content, func(match string) string {
		// Extract variable name (remove ${ and })
		varName := match[2 : len(match)-1]

		// Look up environment variable
		value := os.Getenv(varName)
		if value == "" {
			logrus.WithField("var_name", varName).Warn("Environment variable not found, keeping placeholder")
			return match
		}

		logrus.WithFields(logrus.Fields{
			"var_name": varName,
			"value":    strings.Repeat("*", len(value)), // Mask sensitive values in logs
		}).Debug("Substituted environment variable")

		return value
	})
}

// setDefaults sets default values for optional configuration fields
func setDefaults(cfg *Config) {
	// Server defaults
	if cfg.Server.Version == "" {
		cfg.Server.Version = "1.0.0"
	}

	// Tool defaults
	for i := range cfg.Tools {
		tool := &cfg.Tools[i]

		if tool.Method == "" {
			tool.Method = "GET"
		}

		if tool.ContentType == "" && (tool.Method == "POST" || tool.Method == "PUT" || tool.Method == "PATCH") {
			tool.ContentType = "application/json"
		}

		if tool.Timeout == 0 {
			tool.Timeout = Duration(30 * time.Second)
		}

		if tool.Retries == 0 {
			tool.Retries = 3
		}

		// Set default parameter types
		for j := range tool.Parameters {
			param := &tool.Parameters[j]
			if param.Type == "" {
				param.Type = "string"
			}
		}
	}

	// Security defaults
	if cfg.Security.RateLimit == 0 {
		cfg.Security.RateLimit = 100
	}

	// Runtime defaults
	if cfg.Runtime.MaxConcurrentRequests == 0 {
		cfg.Runtime.MaxConcurrentRequests = 100
	}

	if cfg.Runtime.DefaultTimeout == 0 {
		cfg.Runtime.DefaultTimeout = Duration(30 * time.Second)
	}

	if cfg.Runtime.HealthCheckInterval == 0 {
		cfg.Runtime.HealthCheckInterval = Duration(1 * time.Minute)
	}

	if cfg.Runtime.LogLevel == "" {
		cfg.Runtime.LogLevel = "info"
	}

	if cfg.Runtime.Environment == "" {
		cfg.Runtime.Environment = "development"
	}
}

// validateBusinessRules performs business logic validation
func validateBusinessRules(cfg *Config) error {
	// Validate unique tool names
	toolNames := make(map[string]bool)
	for _, tool := range cfg.Tools {
		if toolNames[tool.Name] {
			return fmt.Errorf("duplicate tool name: %s", tool.Name)
		}
		toolNames[tool.Name] = true
	}

	// Validate unique prompt names
	promptNames := make(map[string]bool)
	for _, prompt := range cfg.Prompts {
		if promptNames[prompt.Name] {
			return fmt.Errorf("duplicate prompt name: %s", prompt.Name)
		}
		promptNames[prompt.Name] = true
	}

	// Validate unique resource URIs
	resourceURIs := make(map[string]bool)
	for _, resource := range cfg.Resources {
		if resourceURIs[resource.URI] {
			return fmt.Errorf("duplicate resource URI: %s", resource.URI)
		}
		resourceURIs[resource.URI] = true

		// Validate resource has content source
		contentSources := 0
		if resource.Content != "" {
			contentSources++
		}
		if resource.FilePath != "" {
			contentSources++
		}
		if resource.URL != "" {
			contentSources++
		}

		if contentSources == 0 {
			return fmt.Errorf("resource %s must have at least one content source (content, file_path, or url)", resource.URI)
		}

		if contentSources > 1 {
			return fmt.Errorf("resource %s can only have one content source", resource.URI)
		}
	}

	// Validate tool authentication
	for _, tool := range cfg.Tools {
		if tool.Auth != nil {
			if err := validateAuthConfig(tool.Auth); err != nil {
				return fmt.Errorf("invalid auth config for tool %s: %w", tool.Name, err)
			}
		}
	}

	return nil
}

// validateAuthConfig validates authentication configuration
func validateAuthConfig(auth *AuthConfig) error {
	switch auth.Type {
	case "bearer":
		if auth.Token == "" && auth.EnvVar == "" {
			return fmt.Errorf("bearer auth requires either token or env_var")
		}
	case "basic":
		if auth.Username == "" || (auth.Password == "" && auth.EnvVar == "") {
			return fmt.Errorf("basic auth requires username and either password or env_var")
		}
	case "api_key":
		if len(auth.Headers) == 0 && auth.EnvVar == "" {
			return fmt.Errorf("api_key auth requires headers or env_var")
		}
	case "custom":
		if len(auth.Headers) == 0 {
			return fmt.Errorf("custom auth requires headers")
		}
	}
	return nil
}

// validateSemVer validates semantic version format
func validateSemVer(fl validator.FieldLevel) bool {
	version := fl.Field().String()
	semverRegex := regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`)
	return semverRegex.MatchString(version)
}
