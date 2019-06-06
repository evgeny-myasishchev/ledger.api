package ledgers

import (
	"net/http"

	"ledger.api/pkg/core/router"
)

func createGetLedgersHandler(svc QueryService) router.ToolkitHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, h router.HandlerToolkit) error {
		result, err := svc.processUserLedgersQuery(r.Context(), &userLedgersQuery{})
		if err != nil {
			return err
		}
		return h.WriteJSON(result)
	}
}

// SetupRoutes - Register ledger related routes
func SetupRoutes(appRouter router.Router, svc QueryService) {

	// TODO
	// server.RequireScopes(createGetLedgersHandler(svc), "read:ledgers")

	appRouter.Handle("GET", "/v2/ledgers", createGetLedgersHandler(svc))
}
