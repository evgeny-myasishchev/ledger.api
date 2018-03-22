package server

import (
	"time"

	"github.com/gin-gonic/gin"
	"ledger.api/pkg/logging"
)

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
