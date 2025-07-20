package config_reader

type ConfigReader interface {
	ReadServiceConfig(serviceName string) (*ServiceConfig, error)
	WatchForChanges(serviceName string, callback func(*ServiceConfig)) error
	StopWatching(serviceName string) error
	GetConfigPath(serviceName string) string
	ValidateConfig(config *ServiceConfig) error
}
