package server

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/jsonapi"
	validator "gopkg.in/go-playground/validator.v9"
)

// HTTPError - standard http error structure
type HTTPError struct {
	Status int

	Errors []*jsonapi.ErrorObject
}

// InternalServerError - return 500 error object
func InternalServerError() *HTTPError {
	return &HTTPError{
		Status: http.StatusInternalServerError,
		Errors: []*jsonapi.ErrorObject{
			{
				Status: strconv.Itoa(http.StatusInternalServerError),
				Title:  http.StatusText(http.StatusInternalServerError),
			},
		},
	}
}

const (
	validationErrDetailsMsg = "Field '%s' validation failed on '%s' tag"
)

// BuildHTTPErrorFromValidationError returns HttpError from validation error
func BuildHTTPErrorFromValidationError(validationErrors *validator.ValidationErrors) *HTTPError {

	ve := *validationErrors
	errLen := len(ve)
	err := HTTPError{
		Status: http.StatusBadRequest,
		Errors: make([]*jsonapi.ErrorObject, errLen),
	}

	for i := 0; i < errLen; i++ {
		fe := ve[i]
		err.Errors[i] = &jsonapi.ErrorObject{
			Status: "400",
			Title:  "Validation error",
			Detail: fmt.Sprintf(validationErrDetailsMsg, fe.Namespace(), fe.Tag()),
		}
	}

	return &err
}

func (e HTTPError) Error() string {
	errorParts := make([]string, len(e.Errors))
	for i, errorPart := range e.Errors {
		errorParts[i] = errorPart.Error()
	}
	return fmt.Sprintf("Error %d: %s", e.Status, strings.Join(errorParts, "; "))
}

// MarshalErrors write error details in a JSON API format
func (e *HTTPError) MarshalErrors(w io.Writer) error {
	if err := jsonapi.MarshalErrors(w, e.Errors); err != nil {
		return err
	}
	return nil
}
