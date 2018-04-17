package server

import uuid "github.com/satori/go.uuid"

// MiddlewareFunc - generic middleware function
type MiddlewareFunc func(*Context, HandlerFunc) (*Response, error)

// NewCallNextMiddleware - creates a middleware that will just call next handler
func NewCallNextMiddleware() MiddlewareFunc {
	return func(ctx *Context, next HandlerFunc) (*Response, error) {
		return next(ctx)
	}
}

// NewRequestIDMiddleware - creates a middleware that will maintain the requestId header
func NewRequestIDMiddleware() MiddlewareFunc {
	return func(ctx *Context, next HandlerFunc) (*Response, error) {
		if requestID := ctx.req.Header.Get("x-request-id"); requestID != "" {
			ctx.requestID = requestID
		} else {
			ctx.requestID = uuid.NewV4().String()
		}
		return next(ctx)
	}
}
