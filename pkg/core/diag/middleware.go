package diag

import (
	"math"
	"net"
	"net/http"
	"runtime"
	"time"

	uuid "github.com/satori/go.uuid"
)

type requestIDMiddlewareCfg struct {
	newUUID func() uuid.UUID
}

type requestIDMiddlewareSetup func(cfg *requestIDMiddlewareCfg)

// NewRequestIDMiddleware - creates a middleware that will maintain the requestId header
func NewRequestIDMiddleware(setup ...requestIDMiddlewareSetup) func(next http.HandlerFunc) http.HandlerFunc {
	cfg := requestIDMiddlewareCfg{newUUID: uuid.NewV4}
	for _, setupFn := range setup {
		setupFn(&cfg)
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			requestID := req.Header.Get("x-request-id")
			if requestID == "" {
				requestID = cfg.newUUID().String()
			}
			nextCtx := ContextWithRequestID(req.Context(), requestID)
			w.Header().Add("x-request-id", requestID)
			next(w, req.WithContext(nextCtx))
		}
	}
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

func (lmw *loggingMiddlewareResponseWrapper) getStatus() int {
	if lmw.status == 0 {
		return 200
	}
	return lmw.status
}

// LogRequestsMiddlewareCfg represents a config for the requests logging middleware
type LogRequestsMiddlewareCfg struct {
	ignorePaths  map[string]bool
	logger       Logger
	runtimeMemMb func() float64
	now          func() time.Time
}

// IgnorePath do not log requests for given path
func (cfg *LogRequestsMiddlewareCfg) IgnorePath(path string) {
	cfg.ignorePaths[path] = true
}

// NewLogRequestsMiddleware - log request start/end
func NewLogRequestsMiddleware(setup ...func(*LogRequestsMiddlewareCfg)) func(next http.HandlerFunc) http.HandlerFunc {
	cfg := LogRequestsMiddlewareCfg{
		ignorePaths: map[string]bool{},
	}
	for _, setupFn := range setup {
		setupFn(&cfg)
	}
	cfg.IgnorePath("/v1/healthcheck/ping")
	if cfg.logger == nil {
		cfg.logger = CreateLogger()
	}
	if cfg.runtimeMemMb == nil {
		cfg.runtimeMemMb = func() float64 {
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			return math.Round(float64(memStats.Alloc)/1024.0/1024.0*1000) / 1000
		}
	}
	if cfg.now == nil {
		cfg.now = time.Now
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			method := req.Method
			path := req.URL.Path

			if _, ok := cfg.ignorePaths[path]; ok {
				next(w, req)
				return
			}

			ip, port, err := net.SplitHostPort(req.RemoteAddr)
			if err != nil {
				cfg.logger.Warn(req.Context(), "Can not parse remote addr: %v", req.RemoteAddr)
				ip = req.RemoteAddr
			}

			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			cfg.logger.
				WithData(MsgData{
					"method":        method,
					"url":           req.URL.RequestURI(),
					"path":          req.URL.Path,
					"userAgent":     req.UserAgent(),
					"headers":       req.Header,
					"query":         req.URL.Query(),
					"remoteAddress": ip,
					"remotePort":    port,
					"memoryUsageMb": cfg.runtimeMemMb(),
				}).
				Info(req.Context(), "BEGIN REQ: %s %s", method, path)

			wrappedWriter := loggingMiddlewareResponseWrapper{
				target: w,
			}
			reqStartedAt := cfg.now()
			next(&wrappedWriter, req)
			reqDuration := cfg.now().Sub(reqStartedAt)

			responseStatus := wrappedWriter.getStatus()

			cfg.logger.
				WithData(MsgData{
					"statusCode":    responseStatus,
					"headers":       w.Header(),
					"duration":      reqDuration.Seconds(),
					"memoryUsageMb": cfg.runtimeMemMb(),
				}).
				Info(req.Context(), "END REQ: %v - %v", responseStatus, path)
		}
	}
}
