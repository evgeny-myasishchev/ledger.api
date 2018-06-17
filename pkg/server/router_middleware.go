package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/jsonapi"
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
		w.Header().Add("x-request-id", requestID)
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

// CreateInitLoggerMiddlewareFunc creates middleware that will init request context
// with a logger instance. Usually should be a very first thing
func CreateInitLoggerMiddlewareFunc(logger logging.Logger) RouterMiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			contextWithLogger := logging.CreateContext(req.Context(), logger)
			requestWithLogger := req.WithContext(contextWithLogger)
			next(w, requestWithLogger)
		}
	}
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
			logger := logging.FromContext(req.Context())
			token, err := validator.ValidateRequest(req)
			if err != nil {
				logger.WithError(err).Error("Token validation failed")
				respondWithErrorStatus(w, http.StatusUnauthorized)
				return
			}

			claims := auth.LedgerClaims{}

			validator.Claims(req, token, &claims)
			if err != nil {
				logger.WithError(err).Error("Failed to get claims")
				respondWithErrorStatus(w, http.StatusUnauthorized)
				return
			}

			nextContext := auth.ContextWithClaims(req.Context(), &claims)
			nextReq := req.WithContext(nextContext)

			next(w, nextReq)
		}
	}
}

// RequireScopes action handler middleware wrapper that will
// verify if scope claim of a token includes scopes provided
func RequireScopes(handler HandlerFunc, scopes ...string) HandlerFunc {
	return HandlerFunc(func(req *http.Request, h *HandlerToolkit) (*Response, error) {
		claims := auth.ClaimsFromContext(req.Context())
		if claims == nil {
			h.Logger.Info("Request has not been initialized with claims, responding with 404")
			return nil, HTTPError{
				Status: http.StatusNotFound,
				Errors: []*jsonapi.ErrorObject{
					{
						Status: strconv.Itoa(http.StatusNotFound),
						Title:  http.StatusText(http.StatusNotFound),
					},
				},
			}
		}

		authorizedScopes := make(map[string]bool)
		for _, authorizedScope := range strings.Split(claims.Scope, ",") {
			authorizedScopes[authorizedScope] = true
		}

		var missingScopes []string
		for _, requiredScope := range scopes {
			if !authorizedScopes[requiredScope] {
				missingScopes = append(missingScopes, requiredScope)
			}
		}

		if len(missingScopes) > 0 {
			h.Logger.Info("Failed to authorize request. Missing scopes: %v", missingScopes)
			return nil, HTTPError{
				Status: http.StatusForbidden,
				Errors: []*jsonapi.ErrorObject{
					{
						Status: strconv.Itoa(http.StatusForbidden),
						Title:  http.StatusText(http.StatusForbidden),
						Detail: fmt.Sprintf("Missing scopes: %v", missingScopes),
					},
				},
			}
		}

		return handler(req, h)
	})
}
