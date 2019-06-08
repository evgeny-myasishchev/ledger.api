package testing

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"ledger.api/pkg/core/router"
)

// NewTestRequest creates an instance of a test request
func NewTestRequest(method string, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(err)
	}
	return req
}

// AssertHTTPErrorResponse asserts if the response is http error response
func AssertHTTPErrorResponse(t *testing.T, expected router.HTTPError, recorder *httptest.ResponseRecorder) {
	var httpError router.HTTPError
	if !JSONUnmarshalReader(t, recorder.Body, &httpError) {
		return
	}
	assert.Equal(t, expected, httpError)
	assert.Equal(t, expected.StatusCode, recorder.Code)
}
