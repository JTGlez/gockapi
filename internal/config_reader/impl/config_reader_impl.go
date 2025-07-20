package impl

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	configReader "github.com/JTGlez/gockapi/internal/config_reader"
)

type ConfigReaderImpl struct {
	BasePath  string
	Validator configReader.ValidatorConfig
	Watchers  map[string]*FileWatcher
	mu        sync.RWMutex
}

type FileWatcher struct {
	ServiceName string
	FilePath    string
	Callback    func(config *configReader.ServiceConfig)
	LastMod     time.Time
	StopChan    chan bool
	Running     bool
}

func NewConfigReader(basePath string) configReader.ConfigReader {
	return &ConfigReaderImpl{
		BasePath: basePath,
		Watchers: make(map[string]*FileWatcher),
	}
}

func (c *ConfigReaderImpl) ReadServiceConfig(serviceName string) (*configReader.ServiceConfig, error) {
	configPath := c.GetConfigPath(serviceName)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist for service %s at path %s", serviceName, configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file for service %s: %w", serviceName, err)
	}

	var config configReader.ServiceConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config for service %s: %w", serviceName, err)
	}

	if err := c.ValidateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed for service %s: %w", serviceName, err)
	}

	log.Printf("Config for %s loaded\n", serviceName)

	return &config, nil
}

func (c *ConfigReaderImpl) WatchForChanges(serviceName string, callback func(*configReader.ServiceConfig)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if existingWatcher, exists := c.Watchers[serviceName]; exists {
		c.stopWatcherInternal(existingWatcher)
	}

	configPath := c.GetConfigPath(serviceName)

	fileInfo, err := os.Stat(configPath)
	if err != nil {
		return fmt.Errorf("cannot watch non-existent file: %s", configPath)
	}

	watcher := &FileWatcher{
		ServiceName: serviceName,
		FilePath:    configPath,
		Callback:    callback,
		LastMod:     fileInfo.ModTime(),
		StopChan:    make(chan bool),
		Running:     true,
	}

	c.Watchers[serviceName] = watcher

	go c.watchFile(watcher)

	return nil
}

func (c *ConfigReaderImpl) StopWatching(serviceName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	watcher, exists := c.Watchers[serviceName]
	if !exists {
		return fmt.Errorf("no watcher found for service %s", serviceName)
	}

	c.stopWatcherInternal(watcher)
	delete(c.Watchers, serviceName)

	return nil
}

func (c *ConfigReaderImpl) GetConfigPath(serviceName string) string {
	return filepath.Join(c.BasePath, serviceName+".json")
}

func (c *ConfigReaderImpl) ValidateConfig(config *configReader.ServiceConfig) error {
	if c.Validator == nil {
		c.Validator = NewConfigValidator()
	}

	return c.Validator.Validate(config)
}

func (c *ConfigReaderImpl) watchFile(watcher *FileWatcher) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-watcher.StopChan:
			return
		case <-ticker.C:
			if !watcher.Running {
				return
			}

			fileInfo, err := os.Stat(watcher.FilePath)
			if err != nil {
				continue
			}

			if fileInfo.ModTime().After(watcher.LastMod) {
				watcher.LastMod = fileInfo.ModTime()

				newConfig, errRead := c.ReadServiceConfig(watcher.ServiceName)
				if errRead != nil {
					log.Printf("Error reloading config for %s: %v\n", watcher.ServiceName, errRead)
					continue
				}

				watcher.Callback(newConfig)
			}
		}
	}
}

func (c *ConfigReaderImpl) stopWatcherInternal(watcher *FileWatcher) {
	if watcher.Running {
		watcher.Running = false
		close(watcher.StopChan)
	}
}
