package server

import "net/http"

// JSON is a shortcup for map[string]interface{}
type JSON map[string]interface{}

// Context - http context interface
type Context interface {
	JSON(code int, obj interface{})
}

// Routes - function to register routes
type Routes func(router Router)

// HandlerFunc - function to handle request
type HandlerFunc func(Context)

// Router - generic router interface
type Router interface {
	GET(relativePath string, handlers HandlerFunc) Router
	POST(relativePath string, handlers HandlerFunc) Router

	RegisterRoutes(routes Routes) Router

	ServeHTTP(w http.ResponseWriter, req *http.Request)

	Run(port int)
}
