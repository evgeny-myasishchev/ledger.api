package app

import "fmt"

type HttpError struct {
	status  int
	message string
}

func (e *HttpError) Error() string {
	return fmt.Sprintf("%d: %s", e.status, e.message)
}
