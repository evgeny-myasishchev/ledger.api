package app

import "ledger.api/pkg/server"

// Routes - Register app routes
func Routes(router *server.Router) {
	router.GET("/v2/healthcheck/ping", func(c *server.Context) (*server.Response, error) {
		return c.R(server.JSON{"message": "pong"}), nil
	})
}
