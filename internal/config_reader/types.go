package config_reader

import "time"

type ServiceConfig struct {
	ServiceName string                    `json:"service_name"`
	Port        int                       `json:"port"`
	Endpoints   map[string]EndpointConfig `json:"endpoints"`
}

type EndpointConfig struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       any               `json:"body,omitempty"`
	Delay      time.Duration     `json:"delay,omitempty"`
}
