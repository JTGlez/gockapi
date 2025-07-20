package config_reader

type ValidatorConfig interface {
	Validate(config *ServiceConfig) error
	ValidateServiceName(serviceName string) error
	ValidatePort(port int) error
	ValidateEndpoint(method, path string, endpoint EndpointConfig) error
}
