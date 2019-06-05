package testing

import (
	"io"
	"net/http"
)

// NewTestRequest creates an instance of a test request
func NewTestRequest(method string, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(err)
	}
	return req
}
