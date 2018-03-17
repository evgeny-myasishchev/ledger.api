package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

type ginContext struct {
	target *gin.Context
}

func (c *ginContext) JSON(code int, obj interface{}) {
	c.target.Render(code, render.JSON{Data: obj})
}

type ginRouter struct {
	engine *gin.Engine
}

func (r *ginRouter) RegisterRoutes(routes Routes) Router {
	routes(r)
	return r
}

func (r *ginRouter) GET(relativePath string, handler HandlerFunc) Router {
	r.engine.GET(relativePath, func(c *gin.Context) {
		handler(&ginContext{target: c})
	})
	return r
}

func (r *ginRouter) POST(relativePath string, handler HandlerFunc) Router {
	r.engine.POST(relativePath, func(c *gin.Context) {
		handler(&ginContext{target: c})
	})
	return r
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
