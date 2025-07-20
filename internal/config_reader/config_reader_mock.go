package config_reader

import (
	"github.com/stretchr/testify/mock"
)

type MockConfigReader struct {
	mock.Mock
}

func (_m *MockConfigReader) GetConfigPath(serviceName string) string {
	ret := _m.Called(serviceName)
	return ret.String(0)
}

func (_m *MockConfigReader) ReadServiceConfig(serviceName string) (*ServiceConfig, error) {
	ret := _m.Called(serviceName)

	var r0 *ServiceConfig
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*ServiceConfig)
	}

	return r0, ret.Error(1)
}

func (_m *MockConfigReader) StopWatching(serviceName string) error {
	ret := _m.Called(serviceName)
	return ret.Error(0)
}

func (_m *MockConfigReader) ValidateConfig(config *ServiceConfig) error {
	ret := _m.Called(config)
	return ret.Error(0)
}

func (_m *MockConfigReader) WatchForChanges(serviceName string, callback func(*ServiceConfig)) error {
	ret := _m.Called(serviceName, callback)
	return ret.Error(0)
}
