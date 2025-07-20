package manager

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"

	configReader "github.com/JTGlez/gockapi/internal/config_reader"
	"github.com/JTGlez/gockapi/internal/config_reader/impl"
	handlers "github.com/JTGlez/gockapi/internal/handlers/response_handler"
	mockServer "github.com/JTGlez/gockapi/internal/server/mock_server"
	portManager "github.com/JTGlez/gockapi/internal/server/port_manager"
)

type MockManager struct {
	servers      map[string]mockServer.MockServer
	configReader configReader.ConfigReader
	portManager  portManager.PortManager
	configPath   string
	running      bool
	mu           sync.RWMutex
}

type ServiceStatus struct {
	ServiceName string            `json:"service_name"`
	Port        int               `json:"port"`
	URL         string            `json:"url"`
	Healthy     bool              `json:"healthy"`
	Status      string            `json:"status"`
	Message     string            `json:"message,omitempty"`
	Details     map[string]string `json:"details,omitempty"`
	LastCheck   string            `json:"last_check,omitempty"`
}

func NewMockManager(configPath string) *MockManager {
	return &MockManager{
		servers:      make(map[string]mockServer.MockServer),
		configReader: impl.NewConfigReader(configPath),
		portManager:  portManager.NewPortManager(),
		configPath:   configPath,
		running:      false,
	}
}

func (m *MockManager) StartAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("mock manager is already running")
	}

	pattern := filepath.Join(m.configPath, "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to list config files: %w", err)
	}

	services := []string{}
	for _, file := range files {
		base := filepath.Base(file)
		serviceName := strings.TrimSuffix(base, filepath.Ext(base))
		services = append(services, serviceName)
	}

	failedServices := []string{}

	for _, serviceName := range services {
		err := m.startServiceInternal(ctx, serviceName)
		if err != nil {
			log.Printf("Failed to start service %s: %v\n", serviceName, err)
			failedServices = append(failedServices, serviceName)
		}
	}

	if len(failedServices) > 0 {
		log.Printf("Warning: %d services failed to start: %v\n", len(failedServices), failedServices)
	}

	if len(m.servers) == 0 {
		return fmt.Errorf("no services could be started")
	}

	m.running = true

	return nil
}

func (m *MockManager) StartService(ctx context.Context, serviceName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.startServiceInternal(ctx, serviceName)
}

func (m *MockManager) startServiceInternal(ctx context.Context, serviceName string) error {
	if _, exists := m.servers[serviceName]; exists {
		return fmt.Errorf("service %s is already running", serviceName)
	}

	cfg, err := m.configReader.ReadServiceConfig(serviceName)
	if err != nil {
		return fmt.Errorf("failed to read config for %s: %w", serviceName, err)
	}

	port, err := m.portManager.AllocatePort(serviceName, cfg.Port)
	if err != nil {
		return fmt.Errorf("failed to allocate port for %s: %w", serviceName, err)
	}

	cfg.Port = port

	responseHandler := handlers.NewResponseHandler()

	httpMockServer := mockServer.NewHTTPMockServer(serviceName, cfg, responseHandler)

	err = httpMockServer.Start(ctx)
	if err != nil {
		m.portManager.ReleasePort(serviceName)
		return fmt.Errorf("failed to start server for %s: %w", serviceName, err)
	}

	m.servers[serviceName] = httpMockServer

	err = m.configReader.WatchForChanges(serviceName, func(newConfig *configReader.ServiceConfig) {
		m.handleConfigChange(serviceName)
	})

	if err != nil {
		log.Printf("Warning: failed to setup hot-reload for %s: %v\n", serviceName, err)
		return err
	}

	return nil
}

func (m *MockManager) StopAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return fmt.Errorf("mock manager is not running")
	}

	errors := []string{}

	for serviceName, mockServer := range m.servers {
		err := mockServer.Stop()
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", serviceName, err))
		}

		m.portManager.ReleasePort(serviceName)

		m.configReader.StopWatching(serviceName)
	}

	m.servers = make(map[string]mockServer.MockServer)
	m.running = false

	if len(errors) > 0 {
		return fmt.Errorf("errors stopping services: %v", errors)
	}

	return nil
}

func (m *MockManager) StopService(serviceName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	server, exists := m.servers[serviceName]
	if !exists {
		return fmt.Errorf("service %s is not running", serviceName)
	}

	err := server.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop server for %s: %w", serviceName, err)
	}

	m.portManager.ReleasePort(serviceName)

	m.configReader.StopWatching(serviceName)

	delete(m.servers, serviceName)

	return nil
}

func (m *MockManager) ReloadAll() error {
	m.mu.RLock()

	serviceNames := make([]string, 0, len(m.servers))
	for serviceName := range m.servers {
		serviceNames = append(serviceNames, serviceName)
	}
	m.mu.RUnlock()

	errors := []string{}

	for _, serviceName := range serviceNames {
		err := m.ReloadService(serviceName)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", serviceName, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors reloading services: %v", errors)
	}

	return nil
}

func (m *MockManager) ReloadService(serviceName string) error {
	m.mu.RLock()
	server, exists := m.servers[serviceName]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("service %s is not running", serviceName)
	}

	newConfig, err := m.configReader.ReadServiceConfig(serviceName)
	if err != nil {
		return fmt.Errorf("failed to read new config for %s: %w", serviceName, err)
	}

	err = server.Reload(newConfig)
	if err != nil {
		return fmt.Errorf("failed to reload server for %s: %w", serviceName, err)
	}

	return nil
}

func (m *MockManager) GetStatus() map[string]ServiceStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]ServiceStatus)

	for serviceName, server := range m.servers {
		healthy := server.IsHealthy()

		serviceStatus := ServiceStatus{
			ServiceName: serviceName,
			Port:        server.GetPort(),
			URL:         server.GetURL(),
			Healthy:     healthy,
			Status:      "running",
		}

		if healthChecker, ok := server.(mockServer.HealthChecker); ok {
			healthStatus := healthChecker.CheckHealth()
			serviceStatus.Message = healthStatus.Message
			serviceStatus.Details = healthStatus.Details
			serviceStatus.LastCheck = healthStatus.Timestamp
		}

		status[serviceName] = serviceStatus
	}

	return status
}

func (m *MockManager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.running
}

func (m *MockManager) GetRunningServices() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	services := make([]string, 0, len(m.servers))
	for serviceName := range m.servers {
		services = append(services, serviceName)
	}

	return services
}

func (m *MockManager) handleConfigChange(serviceName string) {
	log.Printf("🔥 Hot reload: Config change detected for service %s\n", serviceName)

	err := m.ReloadService(serviceName)
	if err != nil {
		log.Printf("❌ Hot reload failed for service %s: %v\n", serviceName, err)
	} else {
		log.Printf("✅ Hot reload successful for service %s\n", serviceName)
	}
}

func (m *MockManager) GetConfigReader() configReader.ConfigReader {
	return m.configReader
}
