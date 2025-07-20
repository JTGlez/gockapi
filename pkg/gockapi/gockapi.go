// Package gockapi provides a simple API for managing mock HTTP servers in tests.
//
// Usage example:
//
//	mgr := gockapi.NewManager("./config-dir")
//	ctx := context.Background()
//
//	err := mgr.StartService(ctx, "my-service")  // Blocks until server is ready
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Test your code immediately - server is guaranteed to be ready
//	resp, _ := http.Get("http://localhost:8080/api/test")
//
//	mgr.StopAll()  // Clean shutdown
package gockapi

import (
	"context"

	"github.com/JTGlez/gockapi/internal/config_reader"
	"github.com/JTGlez/gockapi/internal/manager"
)

type ServiceConfig = config_reader.ServiceConfig

type Manager struct {
	mgr *manager.MockManager
}

// NewManager creates a new mock server manager for attached mode.
// The configPath should point to a directory containing JSON config files.
func NewManager(configPath string) *Manager {
	return &Manager{mgr: manager.NewMockManager(configPath)}
}

// StartAll starts all mock servers from the config directory.
// This method blocks until all servers are ready to accept connections.
// Returns an error if any server fails to start.
func (m *Manager) StartAll(ctx context.Context) error {
	return m.mgr.StartAll(ctx)
}

// StartService starts a single mock server by name.
// This method blocks until the server is ready to accept connections.
// You can immediately make HTTP requests after this method returns successfully.
func (m *Manager) StartService(ctx context.Context, name string) error {
	return m.mgr.StartService(ctx, name)
}

// StopAll stops all running mock servers.
// This provides clean shutdown and port cleanup.
func (m *Manager) StopAll() error {
	return m.mgr.StopAll()
}

// StopService stops a single mock server by name.
func (m *Manager) StopService(name string) error {
	return m.mgr.StopService(name)
}

// GetRunningServices returns the names of currently running services.
func (m *Manager) GetRunningServices() []string {
	return m.mgr.GetRunningServices()
}
