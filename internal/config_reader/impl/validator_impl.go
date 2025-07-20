package impl

import (
	"fmt"
	"strings"

	configReader "github.com/JTGlez/gockapi/internal/config_reader"
)

type ValidatorConfigImpl struct {
	validPortRange struct {
		minPort int
		maxPort int
	}

	validMethods map[string]bool
}

func NewConfigValidator() configReader.ValidatorConfig {
	validator := &ValidatorConfigImpl{
		validMethods: map[string]bool{
			"GET":     true,
			"POST":    true,
			"PUT":     true,
			"DELETE":  true,
			"PATCH":   true,
			"HEAD":    true,
			"OPTIONS": true,
		},
	}

	validator.validPortRange.minPort = 55000
	validator.validPortRange.maxPort = 55999

	return validator
}

func (v ValidatorConfigImpl) Validate(config *configReader.ServiceConfig) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	if err := v.ValidateServiceName(config.ServiceName); err != nil {
		return err
	}

	if err := v.ValidatePort(config.Port); err != nil {
		return err
	}

	if len(config.Endpoints) == 0 {
		fmt.Errorf("at least one endpoint must be configured")
	}

	for endpointKey, endpointConfig := range config.Endpoints {
		parts := strings.SplitN(endpointKey, " ", 2)
		if len(parts) != 2 {
			return fmt.Errorf("endpoint key must be in format 'METHOD /path'")
		}

		method, path := parts[0], parts[1]

		if err := v.ValidateEndpoint(method, path, endpointConfig); err != nil {
			return fmt.Errorf("invalid endpoint %s: %w", endpointKey, err)
		}
	}

	return nil
}

func (v ValidatorConfigImpl) ValidateServiceName(serviceName string) error {
	if serviceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	for _, char := range serviceName {
		if !v.isValidServiceNameChar(char) {
			return fmt.Errorf("service name contains invalid character: %c", char)
		}
	}

	if len(serviceName) > 50 {
		return fmt.Errorf("service name too long (max 50 characters)")
	}

	return nil
}

func (v ValidatorConfigImpl) ValidatePort(port int) error {
	if port <= 0 {
		return fmt.Errorf("port must be positive")
	}

	if port < v.validPortRange.minPort || port > v.validPortRange.maxPort {
		return fmt.Errorf("port must be between %d and %d", v.validPortRange.minPort, v.validPortRange.maxPort)
	}

	return nil
}

func (v ValidatorConfigImpl) ValidateEndpoint(method, path string, endpoint configReader.EndpointConfig) error {
	if !v.validMethods[strings.ToUpper(method)] {
		return fmt.Errorf("invalid HTTP method: %s", method)
	}

	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path must start with /")
	}

	if endpoint.StatusCode <= 0 {
		return fmt.Errorf("status code must be positive")
	}

	if endpoint.StatusCode < 100 || endpoint.StatusCode >= 600 {
		return fmt.Errorf("status code must be between 100 and 599")
	}

	for headerName, headerValue := range endpoint.Headers {
		if headerName == "" {
			return fmt.Errorf("header name cannot be empty")
		}

		if strings.ContainsAny(headerName, " \t\n\r") {
			return fmt.Errorf("header name cannot contain whitespace: %s", headerName)
		}

		_ = headerValue
	}

	return nil
}

func (v ValidatorConfigImpl) isValidServiceNameChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '-' || char == '_'
}
