package ledgers

import (
	"net/http"

	"ledger.api/pkg/server"
)

// CreateRoutes - Register ledger related routes
func CreateRoutes(svc QueryService) server.Routes {
	return func(router *server.Router) {
		router.GET("/v2/ledgers", server.RequireScopes(createGetLedgersHandler(svc), "read:ledgers"))
	}
}

func createGetLedgersHandler(svc QueryService) server.HandlerFunc {
	return func(req *http.Request, h *server.HandlerToolkit) (*server.Response, error) {
		result, err := svc.processUserLedgersQuery(req.Context(), &userLedgersQuery{})
		if err != nil {
			return nil, err
		}
		return h.Response(result), nil
	}
}
