package ledgers

import (
	"net/http"

	"ledger.api/pkg/server"
)

// CreateRoutes - Register ledger related routes
func CreateRoutes() server.Routes {
	return func(router *server.Router) {
		router.GET("/v2/ledgers", server.RequireScopes(handleGetLedgers, "read:ledgers"))
		router.POST("/v2/ledgers", server.RequireScopes(handleCreateLedger, "write:ledgers"))
	}
}

func handleGetLedgers(req *http.Request, h *server.HandlerToolkit) (*server.Response, error) {
	return h.Response(server.JSON{}), nil
}

func handleCreateLedger(req *http.Request, h *server.HandlerToolkit) (*server.Response, error) {
	return h.Response(server.JSON{}), nil
}
