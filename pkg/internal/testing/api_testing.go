package testing

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// NewTestRequest creates an instance of a test request
func NewTestRequest(method string, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(err)
	}
	return req
}

// HTTPErrorPayload is a type that represents HTTPError payload
type HTTPErrorPayload map[string]interface{}

// NewHTTPErrorPayload creates a payload compatible with HTTPError
// used mostly in tests to avoid circular deps
func NewHTTPErrorPayload(code int, status string, message string) HTTPErrorPayload {
	return map[string]interface{}{
		"statusCode": float64(code),
		"error":      status,
		"message":    message,
	}
}

// AssertHTTPErrorResponse asserts if the response is http error response
func AssertHTTPErrorResponse(t *testing.T, expected HTTPErrorPayload, recorder *httptest.ResponseRecorder) {
	if !assert.Equal(t, expected["statusCode"], float64(recorder.Code)) {
		return
	}
	var httpError HTTPErrorPayload
	if !JSONUnmarshalReader(t, recorder.Body, &httpError) {
		return
	}
	assert.Equal(t, expected, httpError)
}
