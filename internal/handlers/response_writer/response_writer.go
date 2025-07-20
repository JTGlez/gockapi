package response_writer

import "net/http"

type ResponseWriter interface {
	WriteJSON(w http.ResponseWriter, statusCode int, headers map[string]string, body interface{}) error
	WriteText(w http.ResponseWriter, statusCode int, headers map[string]string, body string) error
	WriteBytes(w http.ResponseWriter, statusCode int, headers map[string]string, body []byte) error
}
