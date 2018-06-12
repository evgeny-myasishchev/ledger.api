package server

import (
	"context"
	"net/http"

	uuid "github.com/satori/go.uuid"
	"ledger.api/pkg/auth"
	"ledger.api/pkg/logging"
)

type contextKeys string

const requestIDKey contextKeys = "requestID"

// RouterMiddlewareFunc - Generic router middleware interface
type RouterMiddlewareFunc func(next http.HandlerFunc) http.HandlerFunc

// NewRequestIDMiddleware - creates a middleware that will maintain the requestId header
func NewRequestIDMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		logger := logging.FromContext(req.Context())
		requestID := req.Header.Get("x-request-id")
		if requestID == "" {
			requestID = uuid.NewV4().String()
		}
		nextCtx := ContextWithRequestID(req.Context(), requestID)
		nextCtx = logging.CreateContext(nextCtx, logger.WithField("RequestID", requestID))
		next(w, req.WithContext(nextCtx))
	}
}

// ContextWithRequestID - create context with requestID
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDVAlue - returns requestID value taken from context
func RequestIDVAlue(ctx context.Context) string {
	return ctx.Value(requestIDKey).(string)
}

type loggingMiddlewareResponseWrapper struct {
	target http.ResponseWriter
	status int
}

func (lmw *loggingMiddlewareResponseWrapper) Header() http.Header {
	return lmw.target.Header()
}

func (lmw *loggingMiddlewareResponseWrapper) Write(b []byte) (int, error) {
	return lmw.target.Write(b)
}

func (lmw *loggingMiddlewareResponseWrapper) WriteHeader(status int) {
	lmw.target.WriteHeader(status)
	lmw.status = status
}

// NewLoggingMiddleware - log request start/end
func NewLoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		method := req.Method
		path := req.URL.Path

		logger := logging.FromContext(req.Context())

		logger.WithFields(logging.Fields{
			"UserAgent":  req.UserAgent(),
			"RemoteAddr": req.RemoteAddr,
		}).
			Infof("BEGIN REQ: %s %s", method, path)
		wrappedWriter := loggingMiddlewareResponseWrapper{
			target: w,
		}
		// start := time.Now()
		next(&wrappedWriter, req)
		// end := time.Now()
		// duration := end.Sub(start)
		logger.
			// TODO: Optionally response headers
			WithFields(logging.Fields{
				"StatusCode": wrappedWriter.status,
				// "Duration":   duration,
			}).
			Infof("END REQ: %s %s", method, path)
	}
}

// CreateAuthMiddlewareFunc returns auth middleware func that creates auth middleware
func CreateAuthMiddlewareFunc(validator auth.RequestValidator) RouterMiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			validator.ValidateRequest(req)
			token, err := validator.ValidateRequest(req)
			if err != nil {
				// TODO
				panic(err)
			}

			claims := auth.LedgerClaims{}

			validator.Claims(req, token, &claims)
			if err != nil {
				// TODO
				panic(err)
			}

			nextContext := auth.ContextWithClaims(req.Context(), &claims)
			nextReq := req.WithContext(nextContext)

			next(w, nextReq)
		}
	}
}
