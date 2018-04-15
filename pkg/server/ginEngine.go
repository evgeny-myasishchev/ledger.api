package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"ledger.api/pkg/logging"
)

type ginEngine struct {
	ginRouter *gin.Engine
}

func (engine *ginEngine) Handle(method string, path string, handler http.HandlerFunc) {
	engine.ginRouter.Handle(method, path, func(context *gin.Context) {
		handler.ServeHTTP(context.Writer, context.Request)
	})
}

func (engine *ginEngine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	engine.ginRouter.ServeHTTP(w, req)
}

func (engine *ginEngine) Run(port string) error {
	return engine.ginRouter.Run(port)
}

const noRouteErrFmt = `{ "errors": [ { "status": "%v", "title": "%v" } ] }`

var noRouteErrorBody = []byte(fmt.Sprintf(noRouteErrFmt, http.StatusNotFound, http.StatusText(http.StatusNotFound)))

// LoggingMiddleware log request start/end
func LoggingMiddleware(logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.Request.URL.Path

		logger.
			// TODO: Optionally: headers, query
			WithFields(logging.Fields{
				"UserAgent": c.Request.UserAgent(),
				"ClientIP":  c.ClientIP(),
			}).
			Infof("BEGIN REQ: %s %s", method, path)
		start := time.Now()
		c.Next()
		end := time.Now()
		duration := end.Sub(start)
		logger.
			// TODO: Optionally response headers
			WithFields(logging.Fields{
				"StatusCode":    c.Writer.Status(),
				"ContentLength": c.Writer.Size(),
				"Duration":      duration,
			}).
			Infof("END REQ: %s %s", method, path)
	}
}

func createGinEngine(logger logging.Logger) HTTPEngine {

	gin.DisableConsoleColor()
	gin.SetMode(gin.TestMode)
	ginRouter := gin.New()
	ginRouter.Use(LoggingMiddleware(logger))

	ginRouter.NoRoute(func(c *gin.Context) {
		c.Status(http.StatusNotFound)
		c.Writer.Write(noRouteErrorBody)
	})
	return &ginEngine{ginRouter: ginRouter}
}
