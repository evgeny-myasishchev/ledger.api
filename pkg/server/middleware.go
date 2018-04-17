package server

// MiddlewareFunc - generic middleware function
type MiddlewareFunc func(*Context, HandlerFunc) (*Response, error)

// NewCallNextMiddleware - creates a middleware that will just call next handler
func NewCallNextMiddleware() MiddlewareFunc {
	return func(ctx *Context, next HandlerFunc) (*Response, error) {
		return next(ctx)
	}
}
