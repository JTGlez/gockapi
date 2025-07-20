package mock_server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	configReader "github.com/JTGlez/gockapi/internal/config_reader"
	handlers "github.com/JTGlez/gockapi/internal/handlers/response_handler"
)

type MockServerImpl struct {
	serviceName     string
	port            int
	server          *http.Server
	config          *configReader.ServiceConfig
	responseHandler handlers.ResponseHandler
	mu              sync.RWMutex
	running         bool
	healthStatus    HealthStatus
}

func NewHTTPMockServer(serviceName string, cfg *configReader.ServiceConfig, handler handlers.ResponseHandler) MockServer {
	return &MockServerImpl{
		serviceName:     serviceName,
		port:            cfg.Port,
		config:          cfg,
		responseHandler: handler,
		healthStatus: HealthStatus{
			Healthy:   false,
			Service:   serviceName,
			Port:      cfg.Port,
			Message:   "Not started",
			Timestamp: time.Now().Format(time.RFC3339),
		},
	}
}

func (m *MockServerImpl) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("server %s is already running on port %d", m.serviceName, m.port)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", m.handleRequest)

	mux.HandleFunc("/_health", m.handleHealthCheck)

	m.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", m.port),
		Handler: mux,
	}

	go func() {
		if err := m.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			m.mu.Lock()
			m.healthStatus = HealthStatus{
				Healthy:   false,
				Service:   m.serviceName,
				Port:      m.port,
				Message:   fmt.Sprintf("Server failed: %v", err),
				Timestamp: time.Now().Format(time.RFC3339),
			}
			m.running = false
			m.mu.Unlock()
		}
	}()

	m.running = true
	m.healthStatus = HealthStatus{
		Healthy: true,
		Service: m.serviceName,
		Port:    m.port,
		Message: "Server running",
		Details: map[string]string{
			"endpoints": fmt.Sprintf("%d", len(m.config.Endpoints)),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	time.Sleep(100 * time.Millisecond)

	return nil
}

func (m *MockServerImpl) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return fmt.Errorf("server %s is not running", m.serviceName)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.server.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("failed to shutdown server %s: %w", m.serviceName, err)
	}

	m.running = false
	m.healthStatus = HealthStatus{
		Healthy:   false,
		Service:   m.serviceName,
		Port:      m.port,
		Message:   "Server stopped",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return nil
}

func (m *MockServerImpl) Reload(config *configReader.ServiceConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return fmt.Errorf("server %s is not running, cannot reload", m.serviceName)
	}

	if config.ServiceName != m.serviceName {
		return fmt.Errorf("config service name %s does not match server %s", config.ServiceName, m.serviceName)
	}

	if config.Port != m.port {
		return fmt.Errorf("cannot change port from %d to %d via reload, restart required", m.port, config.Port)
	}

	oldConfig := m.config
	m.config = config

	m.healthStatus = HealthStatus{
		Healthy: true,
		Service: m.serviceName,
		Port:    m.port,
		Message: "Configuration reloaded",
		Details: map[string]string{
			"endpoints":      fmt.Sprintf("%d", len(m.config.Endpoints)),
			"last_reload":    time.Now().Format(time.RFC3339),
			"prev_endpoints": fmt.Sprintf("%d", len(oldConfig.Endpoints)),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return nil
}

func (m *MockServerImpl) IsHealthy() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.running && m.healthStatus.Healthy && m.isServerListening()
}

func (m *MockServerImpl) GetURL() string {
	return fmt.Sprintf("http://localhost:%d", m.port)
}

func (m *MockServerImpl) GetServiceName() string {
	return m.serviceName
}

func (m *MockServerImpl) GetPort() int {
	return m.port
}

func (m *MockServerImpl) CheckHealth() HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := m.healthStatus
	status.Timestamp = time.Now().Format(time.RFC3339)

	if m.running && !m.isServerListening() {
		status.Healthy = false
		status.Message = "Server not responding"
	}

	return status
}

func (m *MockServerImpl) GetHealthEndpoint() string {
	return m.GetURL() + "/_health"
}

func (m *MockServerImpl) handleRequest(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	currentConfig := m.config
	m.mu.RUnlock()

	err := m.responseHandler.HandleRequest(w, r, currentConfig)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (m *MockServerImpl) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	healthStatus := m.CheckHealth()

	statusCode := http.StatusOK
	if !healthStatus.Healthy {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	responseHandler := handlers.NewResponseHandler()
	endpointConfig := &configReader.EndpointConfig{
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       healthStatus,
	}

	responseHandler.WriteResponse(w, endpointConfig)
}

func (m *MockServerImpl) isServerListening() bool {
	if !m.running {
		return false
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", m.port), 500*time.Millisecond)
	if err != nil {
		return false
	}
	defer conn.Close()

	return true
}
