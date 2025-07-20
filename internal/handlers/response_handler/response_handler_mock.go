package response_handler

import (
	"net/http"

	configReader "github.com/JTGlez/gockapi/internal/config_reader"
	"github.com/stretchr/testify/mock"
)

type MockResponseHandler struct {
	mock.Mock
}

func (m *MockResponseHandler) HandleRequest(w http.ResponseWriter, r *http.Request, serviceConfig *configReader.ServiceConfig) error {
	args := m.Called(w, r, serviceConfig)
	return args.Error(0)
}

func (m *MockResponseHandler) MatchEndpoint(r *http.Request, endpoints map[string]configReader.EndpointConfig) (*configReader.EndpointConfig, string, error) {
	args := m.Called(r, endpoints)
	return args.Get(0).(*configReader.EndpointConfig), args.String(1), args.Error(2)
}

func (m *MockResponseHandler) WriteResponse(w http.ResponseWriter, endpointConfig *configReader.EndpointConfig) error {
	args := m.Called(w, endpointConfig)
	return args.Error(0)
}

func (m *MockResponseHandler) GetSupportedContentTypes() []string {
	args := m.Called()
	return args.Get(0).([]string)
}
