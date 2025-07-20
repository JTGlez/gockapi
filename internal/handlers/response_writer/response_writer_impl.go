package response_writer

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ResponseWriterImpl struct{}

func (r *ResponseWriterImpl) WriteJSON(w http.ResponseWriter, statusCode int, headers map[string]string, body interface{}) error {
	if str, ok := body.(string); ok {
		_, err := w.Write([]byte(str))
		return err
	}

	if bytes, ok := body.([]byte); ok {
		_, err := w.Write(bytes)
		return err
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	_, err = w.Write(jsonData)

	return err
}

func (r *ResponseWriterImpl) WriteText(w http.ResponseWriter, statusCode int, headers map[string]string, body string) error {
	_, err := w.Write([]byte(body))
	return err
}

func (r *ResponseWriterImpl) WriteBytes(w http.ResponseWriter, statusCode int, headers map[string]string, body []byte) error {
	_, err := w.Write(body)
	return err
}
