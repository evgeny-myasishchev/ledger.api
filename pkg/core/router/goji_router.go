package router

import (
	"encoding/json"
	"net/http"

	"goji.io"
	"goji.io/pat"
)

type gojiHandlerToolkit struct {
	request        *http.Request
	responseWriter http.ResponseWriter
	validator      *structValidator
}

func (h *gojiHandlerToolkit) BindParams() *ParamsBinder {
	return &ParamsBinder{
		req:            h.request,
		validator:      h.validator,
		pathParamValue: pat.Param,
	}
}

func (h *gojiHandlerToolkit) BindPayload(receiver interface{}) error {
	// TODO: Check if can reuse this pattern for params
	if err := json.NewDecoder(h.request.Body).Decode(&receiver); err != nil {
		return err
	}

	// Validator is failing to validate maps so have to ignore explicitly
	_, isMap := receiver.(*map[string]interface{})
	if isMap {
		return nil
	}

	if err := h.validator.validateStruct(h.request.Context(), receiver); err != nil {
		return err
	}

	return nil
}

func (h *gojiHandlerToolkit) WriteJSON(payload interface{}, decorators ...ResponseDecorator) error {

	// This should go first. If we use WithStatus decorator then it will send the header
	// and adding new headers will make no difference
	h.responseWriter.Header().Add("content-type", "application/json")

	for _, decorator := range decorators {
		if err := decorator(h.responseWriter); err != nil {
			return err
		}
	}
	return json.NewEncoder(h.responseWriter).Encode(payload)
}

// WithStatus decorate response with particular http status
func (h *gojiHandlerToolkit) WithStatus(status int) ResponseDecorator {
	return func(w http.ResponseWriter) error {
		w.WriteHeader(status)
		return nil
	}
}

type gojiRouter struct {
	mux       *goji.Mux
	validator *structValidator
}

func (g *gojiRouter) Handle(method string, pattern string, handler ToolkitHandlerFunc) {
	g.mux.HandleFunc(pat.NewWithMethods(pattern, method), func(w http.ResponseWriter, r *http.Request) {
		toolkit := gojiHandlerToolkit{request: r, responseWriter: w, validator: g.validator}
		err := handler(w, r, &toolkit)
		if err != nil {
			logger.WithError(err).Error(r.Context(), "Failed to process request")
			errorResponse := newHTTPErrorFromError(err)
			errorResponse.Send(w)
		}
	})
}

func (g *gojiRouter) Use(mw MiddlewareFunc) {
	g.mux.Use(func(h http.Handler) http.Handler {
		return mw(h.ServeHTTP)
	})
}

func (g *gojiRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}

func createGojiRouter() Router {
	mux := goji.NewMux()
	router := gojiRouter{mux: mux, validator: newStructValidator()}
	return &router
}
