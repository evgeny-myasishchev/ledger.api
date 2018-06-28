package transactions

import (
	"net/http"

	"ledger.api/pkg/server"
)

// CreateRoutes - Register transactions related routes
func CreateRoutes(svc queryService) server.Routes {
	return func(router *server.Router) {
		router.GET(
			// from=:from&to=:to&excludeTags=:excludeTags
			"/v2/ledgers/:ledgerID/transactions/:type/summary",
			// server.RequireScopes(processSummaryQuery, "read:transactions"),
			createSummaryQueryHandler(svc),
		)
	}
}

func createSummaryQueryHandler(svc queryService) server.HandlerFunc {
	return func(req *http.Request, h *server.HandlerToolkit) (*server.Response, error) {
		ledgerID := h.Params.ByName("ledgerID")
		typ := h.Params.ByName("type")
		result, err := svc.processSummaryQuery(req.Context(), &summaryQuery{
			ledgerID: ledgerID,
			typ:      typ,
		})
		if err != nil {
			return nil, err
		}
		return h.JSON(result), nil
	}
}
