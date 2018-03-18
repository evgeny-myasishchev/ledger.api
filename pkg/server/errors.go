package server

import (
	"fmt"
)

// HTTPError - standard http error structure
type HTTPError struct {
	status int
	title  string
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("%d: %s", e.status, e.title)
}

// JSON - Returns a JSON representation of an error
func (e *HTTPError) JSON() JSON {
	return JSON{
		"errors": []JSON{
			{
				"status": e.status,
				"title":  e.title,
			},
		},
	}
}
