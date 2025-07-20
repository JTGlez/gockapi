package response_handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	configReader "github.com/JTGlez/gockapi/internal/config_reader"
	requestMatcher "github.com/JTGlez/gockapi/internal/handlers/request_matcher"
	responseWriter "github.com/JTGlez/gockapi/internal/handlers/response_writer"
)

type ResponseHandlerImpl struct {
	Matcher requestMatcher.RequestMatcher
	Writer  responseWriter.ResponseWriter
}

func NewResponseHandler() ResponseHandler {
	return &ResponseHandlerImpl{
		Matcher: &requestMatcher.RequestMatcherImpl{},
		Writer:  &responseWriter.ResponseWriterImpl{},
	}
}
func (rh *ResponseHandlerImpl) HandleRequest(w http.ResponseWriter, r *http.Request, serviceConfig *configReader.ServiceConfig) error {
	endpointConfig, endpointKey, err := rh.MatchEndpoint(r, serviceConfig.Endpoints)
	if err != nil {
		return fmt.Errorf("failed to match endpoint: %w", err)
	}

	// Si no se encuentra endpoint, retornar 404
	if endpointConfig == nil {
		return rh.writeNotFound(w)
	}

	// Aplicar delay si está configurado
	if endpointConfig.Delay > 0 {
		time.Sleep(endpointConfig.Delay)
	}

	// Escribir respuesta
	err = rh.WriteResponse(w, endpointConfig)
	if err != nil {
		return fmt.Errorf("failed to write response for endpoint %s: %w", endpointKey, err)
	}

	return nil
}

func (rh *ResponseHandlerImpl) MatchEndpoint(r *http.Request, endpoints map[string]configReader.EndpointConfig) (*configReader.EndpointConfig, string, error) {
	for endpointKey, endpointConfig := range endpoints {
		parts := strings.SplitN(endpointKey, " ", 2)
		if len(parts) != 2 {
			continue
		}

		method, pathPattern := parts[0], parts[1]

		matches, pathParams, err := rh.Matcher.Match(r, method, pathPattern)
		if err != nil {
			return nil, "", err
		}

		if matches {
			configCopy := endpointConfig

			_ = pathParams

			return &configCopy, endpointKey, nil
		}
	}

	return nil, "", nil
}

func (rh *ResponseHandlerImpl) WriteResponse(w http.ResponseWriter, endpointConfig *configReader.EndpointConfig) error {
	for key, value := range endpointConfig.Headers {
		w.Header().Set(key, value)
	}

	if endpointConfig.Headers["Content-Type"] == "" && endpointConfig.Body != nil {
		w.Header().Set("Content-Type", "application/json")
	}

	w.WriteHeader(endpointConfig.StatusCode)

	if endpointConfig.Body != nil {
		return rh.Writer.WriteJSON(w, 0, endpointConfig.Headers, endpointConfig.Body)
	}

	return nil
}

func (rh *ResponseHandlerImpl) GetSupportedContentTypes() []string {
	return []string{"application/json", "text/plain"}
}

func (rh *ResponseHandlerImpl) writeNotFound(w http.ResponseWriter) error {
	notFoundConfig := &configReader.EndpointConfig{
		StatusCode: 404,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       map[string]string{"error": "Not Found", "message": "Endpoint not found"},
	}

	return rh.WriteResponse(w, notFoundConfig)
}
