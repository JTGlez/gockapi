package port_manager

import (
	"fmt"
	"net"
	"sync"
)

type PortManager interface {
	AllocatePort(serviceName string, preferredPort int) (int, error)
	ReleasePort(serviceName string) error
	IsPortAvailable(port int) bool
	GetAllocatedPort(serviceName string) (int, bool)
	GetAllocatedPorts() map[string]int
}

type PortManagerImpl struct {
	mu             sync.RWMutex
	allocatedPorts map[string]int // serviceName -> port
	reservedPorts  map[int]bool   // port -> reserved
}

func NewPortManager() PortManager {
	return &PortManagerImpl{
		allocatedPorts: make(map[string]int),
		reservedPorts:  make(map[int]bool),
	}
}

func (p *PortManagerImpl) AllocatePort(serviceName string, preferredPort int) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Validar puerto
	if preferredPort <= 0 {
		return 0, fmt.Errorf("invalid port %d: must be positive", preferredPort)
	}

	// Si el servicio ya tiene un puerto asignado, devolverlo
	if existingPort, exists := p.allocatedPorts[serviceName]; exists {
		return existingPort, nil
	}

	// Si el puerto preferido está disponible, usarlo
	if p.isPortAvailableInternal(preferredPort) {
		p.allocatedPorts[serviceName] = preferredPort
		p.reservedPorts[preferredPort] = true

		return preferredPort, nil
	}

	return 0, fmt.Errorf("preferred port %d is not available for service %s", preferredPort, serviceName)
}

func (p *PortManagerImpl) ReleasePort(serviceName string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	port, exists := p.allocatedPorts[serviceName]
	if !exists {
		return fmt.Errorf("no port allocated for service %s", serviceName)
	}

	delete(p.allocatedPorts, serviceName)
	delete(p.reservedPorts, port)

	return nil
}

func (p *PortManagerImpl) IsPortAvailable(port int) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.isPortAvailableInternal(port)
}

func (p *PortManagerImpl) GetAllocatedPort(serviceName string) (int, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	port, exists := p.allocatedPorts[serviceName]

	return port, exists
}

func (p *PortManagerImpl) GetAllocatedPorts() map[string]int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make(map[string]int)
	for service, port := range p.allocatedPorts {
		result[service] = port
	}

	return result
}

func (p *PortManagerImpl) isPortAvailableInternal(port int) bool {
	if p.reservedPorts[port] {
		return false
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	defer listener.Close()

	return true
}
