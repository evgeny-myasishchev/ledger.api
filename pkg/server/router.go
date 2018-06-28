package server

import (
	"container/list"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/jsonapi"
	validator "gopkg.in/go-playground/validator.v9"
	"ledger.api/pkg/logging"
)

const noRouteErrFmt = `{ "errors": [ { "status": "%v", "title": "%v" } ] }`

var noRouteErrorBody = []byte(fmt.Sprintf(noRouteErrFmt, http.StatusNotFound, http.StatusText(http.StatusNotFound)))

// RequestParams provides unified access for request params
type RequestParams interface {
	ByName(string) string
}

// JSON is a shortcup for map[string]interface{}
type JSON map[string]interface{}

// HandlerToolkit - Collection of various tools to help processing request and build a response
type HandlerToolkit struct {
	validate *validator.Validate
	Logger   logging.Logger
	Params   RequestParams
}

// JSON - Returns JSON response object with status 200
func (h *HandlerToolkit) JSON(obj interface{}) *Response {
	r := &Response{status: 200, json: obj}
	return r
}

// Response - object that holds response data and status
type Response struct {
	json   interface{}
	status int
}

// Status - set custom response status
func (r *Response) Status(status int) *Response {
	r.status = status
	return r
}

// Bind - binds given object to request body (json)
func (h *HandlerToolkit) Bind(req *http.Request, obj interface{}) error {
	err := jsonapi.UnmarshalPayload(req.Body, obj) //TODO: Close req.Body?
	if err != nil {
		return err
	}

	return h.validate.Struct(obj)
}

// Routes - routes registry function
type Routes func(router *Router)

// HandlerFunc - generic route handler function
type HandlerFunc func(req *http.Request, h *HandlerToolkit) (*Response, error)

// Router - http router structure
type Router struct {
	engine     HTTPEngine
	logger     logging.Logger
	validate   *validator.Validate
	middleware list.List
}

// GET - register get route
func (r *Router) GET(relativePath string, handler HandlerFunc) *Router {
	return r.handle("GET", relativePath, handler)
}

// POST - register post route
func (r *Router) POST(relativePath string, handler HandlerFunc) *Router {
	return r.handle("POST", relativePath, handler)
}

func (r *Router) handle(method string, path string, handler HandlerFunc) *Router {
	r.logger.Debugf("Registering route: %v %v", method, path)
	r.engine.Handle(method, path, func(w http.ResponseWriter, req *http.Request) {
		params := req.Context().Value(requestParamsKey).(RequestParams)
		toolkit := HandlerToolkit{
			validate: r.validate,
			Logger:   logging.FromContext(req.Context()),
			Params:   params,
		}
		res, err := handler(req, &toolkit)
		if err != nil {
			toolkit.Logger.WithError(err).Error("Failed to process request")
			respondWithError(w, err)
		} else {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(res.status)
			buffer, err := json.Marshal(res.json)
			if err != nil {
				toolkit.Logger.WithError(err).Error("Failed to marshal json")
				panic(err)
			}
			if _, err := w.Write(buffer); err != nil {
				toolkit.Logger.WithError(err).Error("Failed write buffer")
				panic(err)
			}
		}
	})
	return r
}

func respondWithErrorStatus(w http.ResponseWriter, status int) {
	respondWithError(w, HTTPError{
		Status: status,
		Errors: []*jsonapi.ErrorObject{
			{
				Status: strconv.Itoa(status),
				Title:  http.StatusText(status),
			},
		},
	})
}

func respondWithError(w http.ResponseWriter, err error) {
	httpErr, ok := err.(HTTPError)
	if !ok {
		httpErr = *InternalServerError()
	}
	validationErr, ok := err.(validator.ValidationErrors)
	if ok {
		httpErr = *BuildHTTPErrorFromValidationError(&validationErr)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpErr.Status)
	if err := httpErr.MarshalErrors(w); err != nil {
		panic(err)
	}
}

// HTTPEngine - generic http engine interface that can register routes and serve requests
type HTTPEngine interface {
	Handle(method string, path string, handler http.HandlerFunc)
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

// HTTPApp app structure to register routes and start listening
type HTTPApp struct {
	router *Router
	logger logging.Logger
}

// HTTPAppConfig - config structure for HTTPApp instance
type HTTPAppConfig struct {
	Env    string
	Logger logging.Logger
}

// RegisterRoutes - register app routes
func (app *HTTPApp) RegisterRoutes(routes Routes) *HTTPApp {
	routes(app.router)
	return app
}

// func (app *HTTPApp) ServeHTTP(w http.ResponseWriter, req *http.Request) {
// 	reqWithLogger := req.WithContext(logging.CreateContext(req.Context(), app.logger))
// 	app.router.middleware(w, reqWithLogger)
// }

// Use - Insert another middleware into a call chain
func (app *HTTPApp) Use(middleware RouterMiddlewareFunc) *HTTPApp {
	app.router.middleware.PushBack(middleware)
	return app
}

// UseDefaultMiddleware Initializes default middleware
func (app *HTTPApp) UseDefaultMiddleware() *HTTPApp {
	app.
		Use(CreateInitLoggerMiddlewareFunc(app.logger)).
		Use(NewRequestIDMiddleware).
		Use(NewLoggingMiddleware)
	return app
}

type httpHandler struct {
	target http.HandlerFunc
}

func (handler *httpHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler.target(w, req)
}

// CreateHandler - creates http.Handler instance that can serve requests
// defined by this app
func (app *HTTPApp) CreateHandler() http.Handler {
	target := app.router.engine.ServeHTTP
	for e := app.router.middleware.Back(); e != nil; e = e.Prev() {
		target = e.Value.(RouterMiddlewareFunc)(target)
	}
	return &httpHandler{
		target: target,
	}
}

// CreateHTTPApp - creates an instance of HTTPApp
func CreateHTTPApp(cfg HTTPAppConfig) *HTTPApp {
	logger := cfg.Logger
	if logger == nil {
		logger = logging.NewLogger(cfg.Env)
	}
	logger.Debug("Initializing app router")

	engine := createHTTPRouterEngine(logger)

	router := Router{
		engine:   engine,
		logger:   logger,
		validate: validator.New(),
	}

	httpApp := HTTPApp{
		logger: logger,
		router: &router,
	}

	return httpApp.UseDefaultMiddleware()
}
