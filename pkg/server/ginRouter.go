package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ginContext struct {
}

func (c *ginContext) R(obj JSON) *Response {
	r := &Response{status: 200, json: obj}
	return r
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
		res, err := handler(&ginContext{})
		if err != nil {
			// TODO: Logging here
			// TODO: HttpErrors
			c.JSON(500, JSON{
				"errors": []JSON{
					{
						"status": http.StatusInternalServerError,
						"title":  http.StatusText(http.StatusInternalServerError),
					},
				},
			})
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

// CreateNewRouter - Create new router without any middleware and logging
func CreateNewRouter() Router {
	ginEngine := gin.New()
	router := ginRouter{
		engine: ginEngine,
	}
	return &router
}
