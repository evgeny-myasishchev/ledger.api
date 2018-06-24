package transactions

import (
	"net/http"

	"ledger.api/pkg/server"
)

// CreateRoutes - Register transactions related routes
func CreateRoutes() server.Routes {
	return func(router *server.Router) {
		router.GET(
			"/v2/ledgers/:ledgerID/transactions/summary?type=:type&from=:from&to=:to&excludeTags=:excludeTags",
			server.RequireScopes(handleQuerySummary, "read:transactions"),
		)
	}
}

func handleQuerySummary(req *http.Request, h *server.HandlerToolkit) (*server.Response, error) {
	return h.JSON(server.JSON{}), nil
}
