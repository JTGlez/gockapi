package response_handler

import (
	"net/http"

	configReader "github.com/JTGlez/gockapi/internal/config_reader"
)

type ResponseHandler interface {
	HandleRequest(w http.ResponseWriter, r *http.Request, serviceConfig *configReader.ServiceConfig) error
	MatchEndpoint(r *http.Request, endpoints map[string]configReader.EndpointConfig) (*configReader.EndpointConfig, string, error)
	WriteResponse(w http.ResponseWriter, endpointConfig *configReader.EndpointConfig) error
	GetSupportedContentTypes() []string
}
