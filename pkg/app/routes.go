package app

import "ledger.api/pkg/server"

// Routes - Register app routes
func Routes(router server.Router) {
	router.GET("/v1/healthcheck/ping", func(c server.Context) {
		c.JSON(200, server.JSON{"message": "pong"})
	})
}
