package ledgers

import (
	"ledger.api/pkg/server"
)

// Routes - Register ledger related routes
func Routes(router server.Router) {
	router.GET("/v2/ledgers", func(c server.Context) {
		c.JSON(200, server.JSON{})
	})

	router.POST("/v2/ledgers", func(c server.Context) {
		c.JSON(200, server.JSON{})
	})
}
