package validation

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

// Validator provides validation functionality for MCP server operations
type Validator struct {
	validate *validator.Validate
	logger   *logrus.Logger
}

// New creates a new validator instance
func New() *Validator {
	validate := validator.New()

	// Register custom validation functions
	validate.RegisterValidation("json", validateJSON)
	validate.RegisterValidation("semver", validateSemVer)
	validate.RegisterValidation("endpoint", validateEndpoint)

	return &Validator{
		validate: validate,
		logger:   logrus.New(),
	}
}

// ValidateStruct validates a struct using validation tags
func (v *Validator) ValidateStruct(s interface{}) error {
	if err := v.validate.Struct(s); err != nil {
		return v.formatValidationError(err)
	}
	return nil
}

// ValidateJSON validates that a string is valid JSON
func (v *Validator) ValidateJSON(data string) error {
	var js interface{}
	if err := json.Unmarshal([]byte(data), &js); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

// ValidateParameterType validates that a value matches the expected type
func (v *Validator) ValidateParameterType(value interface{}, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "number":
		switch value.(type) {
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			float32, float64:
			// Valid number types
		case string:
			// Try to parse string as number
			if str := value.(string); str != "" {
				if _, err := strconv.ParseFloat(str, 64); err != nil {
					return fmt.Errorf("string value '%s' is not a valid number", str)
				}
			}
		default:
			return fmt.Errorf("expected number, got %T", value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
	default:
		return fmt.Errorf("unknown parameter type: %s", expectedType)
	}
	return nil
}

// ValidateAPIResponse validates an API response against expected criteria
func (v *Validator) ValidateAPIResponse(response map[string]interface{}, criteria map[string]interface{}) error {
	// Validate required fields
	if requiredFields, ok := criteria["required_fields"].([]string); ok {
		for _, field := range requiredFields {
			if _, exists := response[field]; !exists {
				return fmt.Errorf("required field '%s' missing from response", field)
			}
		}
	}

	// Validate field types
	if fieldTypes, ok := criteria["field_types"].(map[string]string); ok {
		for field, expectedType := range fieldTypes {
			if value, exists := response[field]; exists {
				if err := v.ValidateParameterType(value, expectedType); err != nil {
					return fmt.Errorf("field '%s' validation failed: %w", field, err)
				}
			}
		}
	}

	// Validate field patterns
	if fieldPatterns, ok := criteria["field_patterns"].(map[string]string); ok {
		for field, pattern := range fieldPatterns {
			if value, exists := response[field]; exists {
				if strValue, ok := value.(string); ok {
					matched, err := regexp.MatchString(pattern, strValue)
					if err != nil {
						return fmt.Errorf("invalid pattern for field '%s': %w", field, err)
					}
					if !matched {
						return fmt.Errorf("field '%s' value '%s' does not match pattern '%s'", field, strValue, pattern)
					}
				}
			}
		}
	}

	return nil
}

// ValidateHTTPStatusCode validates that a status code is in the expected range
func (v *Validator) ValidateHTTPStatusCode(statusCode int, expectedCodes []int) error {
	if len(expectedCodes) == 0 {
		// Default: accept 2xx status codes
		if statusCode >= 200 && statusCode < 300 {
			return nil
		}
		return fmt.Errorf("unexpected status code %d, expected 2xx", statusCode)
	}

	for _, expected := range expectedCodes {
		if statusCode == expected {
			return nil
		}
	}

	return fmt.Errorf("unexpected status code %d, expected one of %v", statusCode, expectedCodes)
}

// ValidateURL validates that a string is a valid URL
func (v *Validator) ValidateURL(url string) error {
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// Basic URL validation using regex
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !urlRegex.MatchString(url) {
		return fmt.Errorf("invalid URL format: %s", url)
	}

	return nil
}

// ValidateHTTPMethod validates that a string is a valid HTTP method
func (v *Validator) ValidateHTTPMethod(method string) error {
	validMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

	method = strings.ToUpper(method)
	for _, valid := range validMethods {
		if method == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid HTTP method: %s", method)
}

// ValidateContentType validates that a string is a valid content type
func (v *Validator) ValidateContentType(contentType string) error {
	if contentType == "" {
		return nil // Content type is optional
	}

	validTypes := []string{
		"application/json",
		"application/xml",
		"text/plain",
		"application/x-www-form-urlencoded",
		"multipart/form-data",
	}

	for _, valid := range validTypes {
		if strings.HasPrefix(contentType, valid) {
			return nil
		}
	}

	return fmt.Errorf("unsupported content type: %s", contentType)
}

// formatValidationError formats validation errors into readable messages
func (v *Validator) formatValidationError(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var messages []string

		for _, ve := range validationErrors {
			field := ve.Field()
			tag := ve.Tag()
			param := ve.Param()

			var message string
			switch tag {
			case "required":
				message = fmt.Sprintf("field '%s' is required", field)
			case "min":
				message = fmt.Sprintf("field '%s' must be at least %s characters", field, param)
			case "max":
				message = fmt.Sprintf("field '%s' must be at most %s characters", field, param)
			case "url":
				message = fmt.Sprintf("field '%s' must be a valid URL", field)
			case "oneof":
				message = fmt.Sprintf("field '%s' must be one of: %s", field, param)
			case "semver":
				message = fmt.Sprintf("field '%s' must be a valid semantic version", field)
			case "json":
				message = fmt.Sprintf("field '%s' must be valid JSON", field)
			case "endpoint":
				message = fmt.Sprintf("field '%s' must be a valid API endpoint", field)
			default:
				message = fmt.Sprintf("field '%s' failed validation '%s'", field, tag)
			}

			messages = append(messages, message)
		}

		return fmt.Errorf("validation failed: %s", strings.Join(messages, "; "))
	}

	return err
}

// validateJSON is a custom validator function for JSON strings
func validateJSON(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Empty strings are considered valid
	}

	var js interface{}
	return json.Unmarshal([]byte(value), &js) == nil
}

// validateSemVer is a custom validator function for semantic version strings
func validateSemVer(fl validator.FieldLevel) bool {
	version := fl.Field().String()
	semverRegex := regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`)
	return semverRegex.MatchString(version)
}

// validateEndpoint is a custom validator function for API endpoints
func validateEndpoint(fl validator.FieldLevel) bool {
	endpoint := fl.Field().String()

	// Must be a valid URL starting with http or https
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	return urlRegex.MatchString(endpoint)
}

// SanitizeInput sanitizes input data by removing potentially dangerous content
func (v *Validator) SanitizeInput(data map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})

	for key, value := range data {
		// Sanitize key
		cleanKey := v.sanitizeString(key)

		// Sanitize value based on type
		var cleanValue interface{}
		switch val := value.(type) {
		case string:
			cleanValue = v.sanitizeString(val)
		case map[string]interface{}:
			cleanValue = v.SanitizeInput(val)
		case []interface{}:
			cleanArray := make([]interface{}, len(val))
			for i, item := range val {
				if itemStr, ok := item.(string); ok {
					cleanArray[i] = v.sanitizeString(itemStr)
				} else {
					cleanArray[i] = item
				}
			}
			cleanValue = cleanArray
		default:
			cleanValue = value
		}

		sanitized[cleanKey] = cleanValue
	}

	return sanitized
}

// sanitizeString removes potentially dangerous characters from strings
func (v *Validator) sanitizeString(s string) string {
	// Remove null bytes
	s = strings.ReplaceAll(s, "\x00", "")

	// Remove control characters except common whitespace
	var result strings.Builder
	for _, r := range s {
		if r >= 32 || r == '\t' || r == '\n' || r == '\r' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// GetValidationErrorType returns the type of validation error for better error handling
func (v *Validator) GetValidationErrorType(err error) string {
	if err == nil {
		return "none"
	}

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		if len(validationErrors) > 0 {
			return validationErrors[0].Tag()
		}
	}

	// Check for common error patterns
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "required"):
		return "required"
	case strings.Contains(errStr, "type"):
		return "type"
	case strings.Contains(errStr, "format"):
		return "format"
	case strings.Contains(errStr, "range"):
		return "range"
	default:
		return "unknown"
	}
}
