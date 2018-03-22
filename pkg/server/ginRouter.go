package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/jsonapi"
	"ledger.api/pkg/logging"
)

type ginContext struct {
	target *gin.Context
}

func (c *ginContext) R(obj JSON) *Response {
	r := &Response{status: 200, json: obj}
	return r
}

func (c *ginContext) Bind(obj interface{}) error {
	return jsonapi.UnmarshalPayload(c.target.Request.Body, obj)
}

type ginRouter struct {
	engine *gin.Engine
}

func (r *ginRouter) RegisterRoutes(routes Routes) Router {
	routes(r)
	return r
}

func (r *ginRouter) handle(httpMethod string, relativePath string, handler HandlerFunc) Router {
	r.engine.Handle(httpMethod, relativePath, func(c *gin.Context) {
		res, err := handler(&ginContext{target: c})
		if err != nil {

			httpErr, ok := err.(HTTPError)
			if !ok {
				httpErr = HTTPError{
					status: http.StatusInternalServerError,
					title:  http.StatusText(http.StatusInternalServerError),
				}
			}
			// TODO: Logging here
			c.JSON(httpErr.status, httpErr.JSON())
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
		engine: ginEngine,
	}
	return &router
}
