package app

import "github.com/gin-gonic/gin"

// RegisterRoutes - Register app routes
func RegisterRoutes(router *gin.Engine) {
	router.GET("/v1/healthcheck/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
}
