package server

import (
	"container/list"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/jsonapi"
	validator "gopkg.in/go-playground/validator.v9"
	"ledger.api/pkg/logging"
)

const noRouteErrFmt = `{ "errors": [ { "status": "%v", "title": "%v" } ] }`

var noRouteErrorBody = []byte(fmt.Sprintf(noRouteErrFmt, http.StatusNotFound, http.StatusText(http.StatusNotFound)))

// JSON is a shortcup for map[string]interface{}
type JSON map[string]interface{}

// HandlerToolkit - Collection of various tools to help processing request and build a response
type HandlerToolkit struct {
	validate *validator.Validate
	Logger   logging.Logger
}

// JSON - Returns JSON response object with status 200
func (h *HandlerToolkit) JSON(obj JSON) *Response {
	r := &Response{status: 200, json: obj}
	return r
}

// Response - object that holds response data and status
type Response struct {
	json   JSON
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
		toolkit := HandlerToolkit{
			validate: r.validate,
			Logger:   logging.FromContext(req.Context()),
		}
		res, err := handler(req, &toolkit)
		if err != nil {
			httpErr, ok := err.(HTTPError)
			if !ok {
				httpErr = *InternalServerError()
			}
			validationErr, ok := err.(validator.ValidationErrors)
			if ok {
				httpErr = *BuildHTTPErrorFromValidationError(&validationErr)
			}
			w.WriteHeader(httpErr.Status)
			if err := httpErr.MarshalErrors(w); err != nil {
				panic(err)
			}
		} else {
			w.WriteHeader(res.status)
			buffer, err := json.Marshal(res.json)
			if err != nil {
				panic(err)
			}
			if _, err := w.Write(buffer); err != nil {
				panic(err)
			}
		}
	})
	return r
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
		Use(func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, req *http.Request) {
				contextWithLogger := logging.CreateContext(req.Context(), app.logger)
				requestWithLogger := req.WithContext(contextWithLogger)
				next(w, requestWithLogger)
			}
		}).
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
