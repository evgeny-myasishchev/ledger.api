package server

import (
	"encoding/json"
	"net/http"

	"github.com/google/jsonapi"
	validator "gopkg.in/go-playground/validator.v9"
	"ledger.api/pkg/logging"
)

type Context2 struct {
	req      *http.Request
	validate *validator.Validate

	Logger logging.Logger
}

func (c *Context2) R(obj JSON) *Response {
	r := &Response{status: 200, json: obj}
	return r
}

func (c *Context2) Bind(obj interface{}) error {
	err := jsonapi.UnmarshalPayload(c.req.Body, obj) //TODO: Close req.Body?
	if err != nil {
		return err
	}

	return c.validate.Struct(obj)
}

type Routes2 func(router *Router2)

type HandlerFunc2 func(*Context2) (*Response, error)

type Router2 struct {
	engine   HttpEngine
	logger   logging.Logger
	validate *validator.Validate
}

func (r *Router2) GET(relativePath string, handler HandlerFunc2) *Router2 {
	return r.handle("GET", relativePath, handler)
}

func (r *Router2) POST(relativePath string, handler HandlerFunc2) *Router2 {
	return r.handle("POST", relativePath, handler)
}

func (r *Router2) handle(method string, path string, handler HandlerFunc2) *Router2 {
	r.logger.Debugf("Registering route: %v %v", method, path)
	r.engine.Handle(method, path, func(w http.ResponseWriter, req *http.Request) {
		res, err := handler(&Context2{
			validate: r.validate,
			Logger:   r.logger, //TODO: Child logger with RequestID
		})
		if err != nil {
			r.logger.WithError(err).Error("Failed to process request")
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
			enc := json.NewEncoder(w)
			if err := enc.Encode(res.json); err != nil {
				panic(err) //TODO: Perhaps not to fail hard here
			}
		}
	})
	return r
}

type HttpEngine interface {
	Handle(method string, path string, handler http.HandlerFunc)
}

type HttpApp struct {
	router *Router2

	// RegisterRoutes(routes Routes2) *Router2
	//
	// ServeHTTP(w http.ResponseWriter, req *http.Request)
	//
	// Run(port int)
}

func (app *HttpApp) RegisterRoutes(routes Routes2) *HttpApp {
	routes(app.router)
	return app
}

func CreateHttpApp() *HttpApp {
	logger := logging.NewTestLogger()
	logger.Debug("Initializing test router")

	router := Router2{
		logger: logger,
	}

	httpApp := HttpApp{
		router: &router,
	}

	// ginEngine := gin.New()
	// ginEngine.Use(LoggingMiddleware(logger))
	// router := ginRouter{
	// 	engine:   ginEngine,
	// 	logger:   logger,
	// 	validate: validator.New(),
	// }
	//
	// ginEngine.NoRoute(func(c *gin.Context) {
	// 	c.Status(http.StatusNotFound)
	// 	c.Writer.Write(noRouteErrorBody)
	// })
	return &httpApp
}
