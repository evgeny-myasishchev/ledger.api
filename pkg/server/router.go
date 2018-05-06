package server

import (
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

// Context request context structure
type Context struct {
	req       *http.Request
	validate  *validator.Validate
	requestID string

	Logger logging.Logger
}

// R returns default response structure
func (c *Context) R(obj JSON) *Response {
	r := &Response{status: 200, json: obj}
	return r
}

// Response - object that holds response data and status
type Response struct {
	json   JSON
	status int
}

// S - set custom response status
func (r *Response) S(status int) *Response {
	r.status = status
	return r
}

// Bind - binds given object to request body (json)
func (c *Context) Bind(obj interface{}) error {
	err := jsonapi.UnmarshalPayload(c.req.Body, obj) //TODO: Close req.Body?
	if err != nil {
		return err
	}

	return c.validate.Struct(obj)
}

// Routes - routes registry function
type Routes func(router *Router)

// HandlerFunc - generic route handler function
type HandlerFunc func(*Context) (*Response, error)

// Router - http router structure
type Router struct {
	engine     HTTPEngine
	logger     logging.Logger
	validate   *validator.Validate
	middleware MiddlewareFunc
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
		context := Context{
			req:      req,
			validate: r.validate,
			Logger:   r.logger, //TODO: Child logger with RequestID
		}
		res, err := r.middleware(&context, handler)
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
	Run(port string) error
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

func (app *HTTPApp) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	app.router.engine.ServeHTTP(w, req)
}

// Run - start server for given port
func (app *HTTPApp) Run(port int) {
	app.logger.Infof("Starting server on port: %v", port)
	err := app.router.engine.Run(fmt.Sprintf(":%v", port))
	if err != nil {
		panic(err)
	}
}

// Use - Insert another middleware into a call chain
func (app *HTTPApp) Use(middleware MiddlewareFunc) *HTTPApp {
	head := app.router.middleware
	app.router.middleware = func(context *Context, next HandlerFunc) (*Response, error) {
		return head(context, func(ctx *Context) (*Response, error) {
			return middleware(ctx, next)
		})
	}
	return app
}

// UseDefaultMiddleware Initializes default middleware
func (app *HTTPApp) UseDefaultMiddleware() *HTTPApp {
	// TODO: Restore after reworking middleware
	// app.
	// Use(NewRequestIDMiddleware()).
	// Use(NewLoggingMiddleware())
	return app
}

// CreateHTTPApp - creates an instance of HTTPApp
func CreateHTTPApp(cfg HTTPAppConfig) *HTTPApp {
	logger := cfg.Logger
	if logger == nil {
		logger = logging.NewLogger(cfg.Env)
	}
	logger.Debug("Initializing test router")

	router := Router{
		engine:     createHTTPRouterEngine(logger),
		logger:     logger,
		validate:   validator.New(),
		middleware: NewCallNextMiddleware(),
	}

	httpApp := HTTPApp{
		logger: logger,
		router: &router,
	}

	return httpApp.UseDefaultMiddleware()
}
