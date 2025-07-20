package config_reader

import (
	"github.com/stretchr/testify/mock"
)

type MockValidatorConfig struct {
	mock.Mock
}

func (m *MockValidatorConfig) Validate(config *ServiceConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockValidatorConfig) ValidateServiceName(serviceName string) error {
	args := m.Called(serviceName)
	return args.Error(0)
}

func (m *MockValidatorConfig) ValidatePort(port int) error {
	args := m.Called(port)
	return args.Error(0)
}

func (m *MockValidatorConfig) ValidateEndpoint(method, path string, endpoint EndpointConfig) error {
	args := m.Called(method, path, endpoint)
	return args.Error(0)
}
