package response_writer

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type MockResponseWriter struct {
	mock.Mock
}

func (m *MockResponseWriter) WriteJSON(w http.ResponseWriter, statusCode int, headers map[string]string, body interface{}) error {
	args := m.Called(w, statusCode, headers, body)
	return args.Error(0)
}

func (m *MockResponseWriter) WriteText(w http.ResponseWriter, statusCode int, headers map[string]string, body string) error {
	args := m.Called(w, statusCode, headers, body)
	return args.Error(0)
}

func (m *MockResponseWriter) WriteBytes(w http.ResponseWriter, statusCode int, headers map[string]string, body []byte) error {
	args := m.Called(w, statusCode, headers, body)
	return args.Error(0)
}
