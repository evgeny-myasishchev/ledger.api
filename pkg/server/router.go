package server

import "net/http"

// JSON is a shortcup for map[string]interface{}
type JSON map[string]interface{}

// Context - http context interface
type Context interface {
	R(json JSON) *Response
	Bind(obj interface{}) error
}

// Routes - function to register routes
type Routes func(router Router)

// HandlerFunc - function to handle request
type HandlerFunc func(Context) (*Response, error)

// Router - generic router interface
type Router interface {
	GET(relativePath string, handlers HandlerFunc) Router
	POST(relativePath string, handlers HandlerFunc) Router

	RegisterRoutes(routes Routes) Router

	ServeHTTP(w http.ResponseWriter, req *http.Request)

	Run(port int)
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
