package ledgers

import (
	"ledger.api/pkg/server"
)

// CreateRoutes - Register ledger related routes
func CreateRoutes(ledgerSvc Service) server.Routes {
	return func(router *server.Router) {
		router.GET("/v2/ledgers", func(c *server.Context) (*server.Response, error) {
			return c.R(server.JSON{}), nil
		})

		router.POST("/v2/ledgers", func(c *server.Context) (*server.Response, error) {
			return c.R(server.JSON{}), nil
		})
	}
}
