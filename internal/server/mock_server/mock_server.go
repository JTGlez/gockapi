package mock_server

import (
	"context"

	configReader "github.com/JTGlez/gockapi/internal/config_reader"
)

type MockServer interface {
	Start(ctx context.Context) error
	Stop() error
	Reload(config *configReader.ServiceConfig) error
	IsHealthy() bool
	GetURL() string
	GetServiceName() string
	GetPort() int
}

type HealthChecker interface {
	CheckHealth() HealthStatus
	GetHealthEndpoint() string
}

type HealthStatus struct {
	Healthy   bool              `json:"healthy"`
	Service   string            `json:"service"`
	Port      int               `json:"port,omitempty"`
	Message   string            `json:"message,omitempty"`
	Details   map[string]string `json:"details,omitempty"`
	Timestamp string            `json:"timestamp"`
}
