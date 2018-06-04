package ledgers

import (
	"net/http"

	"ledger.api/pkg/server"
)

// CreateRoutes - Register ledger related routes
func CreateRoutes(ledgerSvc Service) server.Routes {
	return func(router *server.Router) {
		router.GET("/v2/ledgers", handleGetLedgers)
		router.POST("/v2/ledgers", handleCreateLedger)
	}
}

func handleGetLedgers(req *http.Request, h *server.HandlerToolkit) (*server.Response, error) {
	return h.JSON(server.JSON{}), nil
}

func handleCreateLedger(req *http.Request, h *server.HandlerToolkit) (*server.Response, error) {
	return h.JSON(server.JSON{}), nil
}
