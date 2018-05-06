package server

import (
	uuid "github.com/satori/go.uuid"
)

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
		ctx.Logger = ctx.Logger.WithField("RequestID", ctx.requestID)
		return next(ctx)
	}
}

// NewLoggingMiddleware - log request start/end
// func NewLoggingMiddleware() MiddlewareFunc {
// 	return func(ctx *Context, next HandlerFunc) (*Response, error) {
//
// 		method := ctx.req.Method
// 		path := ctx.req.URL.Path
//
// 		logger := ctx.Logger
//
// 		logger.WithFields(logging.Fields{
// 			"UserAgent":  ctx.req.UserAgent(),
// 			"RemoteAddr": ctx.req.RemoteAddr,
// 		}).
// 			Infof("BEGIN REQ: %s %s", method, path)
// 		start := time.Now()
// 		res, err := next(ctx)
// 		if err != nil {
// 			logger.WithError(err).Error("Failed to process request")
// 		}
// 		end := time.Now()
// 		duration := end.Sub(start)
// 		logger.
// 			// TODO: Optionally response headers
// 			WithFields(logging.Fields{
// 				"StatusCode": res.status,
// 				"Duration":   duration,
// 			}).
// 			Infof("END REQ: %s %s", method, path)
// 		return res, err
// 	}
// }
