package server

import (
	"github.com/gin-gonic/gin"
	"ledger.api/pkg/logging"
)

// LoggingMiddleware log request start/end
func LoggingMiddleware(logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		// start := time.Now()
		// path := c.Request.URL.Path
		// raw := c.Request.URL.RawQuery

		// Process request
		c.Next()
	}
}
