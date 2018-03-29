package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/jsonapi"
	validator "gopkg.in/go-playground/validator.v9"
	"ledger.api/pkg/logging"
)

type ginContext struct {
	target   *gin.Context
	validate *validator.Validate
	logger   logging.Logger
}

func (c *ginContext) Logger() logging.Logger {
	return c.logger
}

func (c *ginContext) R(obj JSON) *Response {
	r := &Response{status: 200, json: obj}
	return r
}

func (c *ginContext) Bind(obj interface{}) error {
	err := jsonapi.UnmarshalPayload(c.target.Request.Body, obj)
	if err != nil {
		return err
	}

	return c.validate.Struct(obj)
}

type ginRouter struct {
	engine   *gin.Engine
	logger   logging.Logger
	validate *validator.Validate
}

func (r *ginRouter) RegisterRoutes(routes Routes) Router {
	routes(r)
	return r
}

func (r *ginRouter) handle(httpMethod string, relativePath string, handler HandlerFunc) Router {
	r.engine.Handle(httpMethod, relativePath, func(c *gin.Context) {
		res, err := handler(&ginContext{
			target:   c,
			validate: r.validate,
			logger:   r.logger, //TODO: Child logger with RequestID
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
			c.Status(httpErr.Status)
			if err := httpErr.MarshalErrors(c.Writer); err != nil {
				panic(err)
			}
		} else {
			c.JSON(res.status, res.json)
		}
	})
	return r
}

func (r *ginRouter) GET(relativePath string, handler HandlerFunc) Router {
	return r.handle("GET", relativePath, handler)
}

func (r *ginRouter) POST(relativePath string, handler HandlerFunc) Router {
	return r.handle("POST", relativePath, handler)
}

func (r *ginRouter) Run(port int) {
	err := r.engine.Run(fmt.Sprintf(":%v", port))
	if err != nil {
		panic(err)
	}
}

func (r *ginRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.engine.ServeHTTP(w, req)
}

// CreateDefaultRouter - Create default router with middlewares and logging
func CreateDefaultRouter() Router {
	ginEngine := gin.Default()
	router := ginRouter{
		engine: ginEngine,
	}
	return &router
}

// CreateTestRouter - Create new router that can be used for tests
func CreateTestRouter() Router {
	logger := logging.NewTestLogger()
	logger.Debug("Initializing test router")
	gin.DisableConsoleColor()
	gin.SetMode(gin.TestMode)
	ginEngine := gin.New()
	ginEngine.Use(LoggingMiddleware(logger))
	router := ginRouter{
		engine:   ginEngine,
		logger:   logger,
		validate: validator.New(),
	}
	return &router
}
