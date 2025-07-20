# gockapi

A powerful tool for spinning up mock HTTP servers from JSON configurations. gockapi supports two modes of operation: **detached mode** (CLI tool) for standalone mock servers and **attached mode** (Go library) for programmatic control in tests.

---

## Features

- 🚀 **Two Operation Modes**: CLI tool for standalone servers + Go library for tests
- 📁 **JSON Configuration**: Define endpoints, responses, and ports via simple JSON files  
- 🔄 **Flexible Deployment**: Start services individually or all at once
- 🎯 **Stateless Management**: CLI discovers and manages services by port scanning
- ⚡ **Zero Wait Time**: Attached mode blocks until servers are ready (no manual delays)
- 🧹 **Clean Shutdown**: Reliable process management and port cleanup

---

## Installation

### Build and Install Locally (Detached mode)

```bash
# Build the binary and place it in your Go bin directory
go build -o ~/go/bin/gockapi ./cmd/gockapi

# Or use go install from the project root
go install ./cmd/gockapi
```

### Add to PATH

Make sure your Go bin directory is in your `PATH`:

```bash
export PATH=$PATH:~/go/bin
```

To make this permanent, add the above line to your `~/.bashrc` or `~/.profile`.

---

### As dependency (Attached mode)

```bash
go get github.com/JTGlez/gockapi@v0.0.1
```

## Configuration

Create JSON configuration files for your mock services:

**Example: `my-configs/userService.json`**
```json
{
  "service_name": "userService",
  "port": 8001,
  "endpoints": {
    "GET /api/users": {
      "status_code": 200,
      "headers": {
        "Content-Type": "application/json"
      },
      "body": {
        "users": [
          {"id": 1, "name": "John Doe"},
          {"id": 2, "name": "Jane Smith"}
        ]
      }
    },
    "POST /api/users": {
      "status_code": 201,
      "headers": {
        "Content-Type": "application/json"
      },
      "body": {
        "message": "User created successfully"
      }
    }
  }
}
```

---

## Detached Mode (CLI Tool)

Use gockapi as a standalone CLI tool for spinning up mock servers that run independently of your application.

### Quick Start

```bash
# Check status of all configured services
gockapi --config-path ./my-configs status

# Start a specific service in the background
gockapi --config-path ./my-configs start userService &

# Test your mock server
curl http://localhost:8001/api/users

# Stop the service
gockapi --config-path ./my-configs stop-all
```

### CLI Commands

| Command | Description |
|---------|-------------|
| `start-all` | Start all services in config directory (single process) |
| `start <service>` | Start a specific service by name |
| `stop-all` | Stop all running services (stateless port scanning) |
| `stop <service>` | Stop a specific service |
| `status` | Show status of all configured services |

### Deployment Patterns

#### Pattern 1: Single Process (All Services Together)
```bash
# All services run in one process
gockapi --config-path ./my-configs start-all &

# Or start multiple specific services together
gockapi --config-path ./my-configs start userService paymentService &
```

#### Pattern 2: Separate Processes (Independent Services)
```bash
# Each service runs in its own process
gockapi --config-path ./my-configs start userService &
gockapi --config-path ./my-configs start paymentService &
gockapi --config-path ./my-configs start notificationService &

# Stop individual services without affecting others
gockapi --config-path ./my-configs stop userService
```

#### Pattern 3: Stateless Management
```bash
# Stop all services regardless of how they were started
gockapi --config-path ./my-configs stop-all

# Works even if services were started in different shells or scripts
# The tool scans ports from all config files and kills matching processes
```

### Example Workflow

```bash
# 1. Create your config directory
mkdir my-configs

# 2. Add service configurations (see Configuration section above)

# 3. Start services
gockapi --config-path ./my-configs start userService &
gockapi --config-path ./my-configs start paymentService &

# 4. Verify services are running
gockapi --config-path ./my-configs status

# 5. Test your services
curl http://localhost:8001/api/users
curl http://localhost:8002/api/payments

# 6. Clean shutdown
gockapi --config-path ./my-configs stop-all
```

---

## Attached Mode (Go Library)

Use gockapi as a Go library for programmatic control of mock servers in your tests. Perfect for integration testing where you need reliable, fast mock server lifecycle management.

### Installation

```bash
go get github.com/JTGlez/gockapi/pkg/gockapi
```

### Quick Start

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "log"
    
    "github.com/JTGlez/gockapi/pkg/gockapi"
)

func main() {
    // 1. Create manager pointing to your config directory
    mgr := gockapi.NewManager("./my-configs")
    defer mgr.StopAll() // Always clean up

    // 2. Start a service - blocks until ready, no sleep needed!
    ctx := context.Background()
    err := mgr.StartService(ctx, "userService")
    if err != nil {
        log.Fatal(err)
    }

    // 3. Test immediately - server is guaranteed ready
    resp, err := http.Get("http://localhost:8001/api/users")
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()
    
    fmt.Printf("Status: %d\n", resp.StatusCode)
    // 4. That's it! StopAll in defer handles cleanup
}
```

### API Reference

```go
type Manager struct {
    // Internal fields
}

// NewManager creates a new mock server manager
func NewManager(configPath string) *Manager

// StartAll starts all mock servers from the config directory
// Blocks until all servers are ready to accept connections
func (m *Manager) StartAll(ctx context.Context) error

// StartService starts a single mock server by name
// Blocks until the server is ready - you can immediately make requests
func (m *Manager) StartService(ctx context.Context, name string) error

// StopAll stops all running mock servers with clean shutdown
func (m *Manager) StopAll() error

// StopService stops a single mock server by name
func (m *Manager) StopService(name string) error

// GetRunningServices returns the names of currently running services
func (m *Manager) GetRunningServices() []string
```

### Testing Patterns

#### Pattern 1: Single Service Test
```go
func TestUserAPI(t *testing.T) {
    mgr := gockapi.NewManager("./test-configs")
    defer mgr.StopAll()

    ctx := context.Background()
    err := mgr.StartService(ctx, "userService")
    if err != nil {
        t.Fatalf("Failed to start service: %v", err)
    }

    // Test your code that depends on the mock service
    resp, err := http.Get("http://localhost:8001/api/users")
    if err != nil {
        t.Fatalf("Request failed: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        t.Errorf("Expected 200, got %d", resp.StatusCode)
    }
}
```

#### Pattern 2: Multiple Services Test
```go
func TestFullIntegration(t *testing.T) {
    mgr := gockapi.NewManager("./test-configs")
    defer mgr.StopAll()

    ctx := context.Background()
    
    // Start all services at once
    err := mgr.StartAll(ctx)
    if err != nil {
        t.Fatalf("Failed to start services: %v", err)
    }

    // Test complex scenarios involving multiple services
    // All services are guaranteed to be ready
    testUserWorkflow(t)
    testPaymentWorkflow(t)
    testNotificationWorkflow(t)
}
```

#### Pattern 3: Table-Driven Tests
```go
func TestMockServices(t *testing.T) {
    tests := []struct {
        name     string
        service  string
        endpoint string
        expect   int
    }{
        {"User API", "userService", "http://localhost:8001/api/users", 200},
        {"Payment API", "paymentService", "http://localhost:8002/api/payments", 200},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            mgr := gockapi.NewManager("./test-configs")
            defer mgr.StopAll()

            ctx := context.Background()
            err := mgr.StartService(ctx, tc.service)
            if err != nil {
                t.Fatalf("Failed to start %s: %v", tc.service, err)
            }

            resp, err := http.Get(tc.endpoint)
            if err != nil {
                t.Fatalf("Request to %s failed: %v", tc.endpoint, err)
            }
            defer resp.Body.Close()

            if resp.StatusCode != tc.expect {
                t.Errorf("Expected %d, got %d", tc.expect, resp.StatusCode)
            }
        })
    }
}
```

## Advanced Configuration

### Multiple Endpoints Per Service

```json
{
  "service_name": "apiGateway",
  "port": 8000,
  "endpoints": {
    "GET /health": {
      "status_code": 200,
      "body": {"status": "healthy"}
    },
    "GET /api/v1/users/:id": {
      "status_code": 200,
      "body": {"id": 1, "name": "John Doe"}
    },
    "POST /api/v1/users": {
      "status_code": 201,
      "headers": {"Location": "/api/v1/users/123"},
      "body": {"message": "User created"}
    },
    "DELETE /api/v1/users/:id": {
      "status_code": 204,
      "body": ""
    }
  }
}
```

### Error Responses

```json
{
  "service_name": "errorService",
  "port": 8080,
  "endpoints": {
    "GET /api/error": {
      "status_code": 500,
      "headers": {
        "Content-Type": "application/json"
      },
      "body": {
        "error": "Internal server error",
        "code": "INTERNAL_ERROR"
      }
    },
    "GET /api/notfound": {
      "status_code": 404,
      "body": {
        "error": "Resource not found"
      }
    }
  }
}
```

---

## License

MIT
