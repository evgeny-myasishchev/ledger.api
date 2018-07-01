package app

import (
	"net/http"

	"ledger.api/pkg/server"
)

// Routes - Register app routes
func Routes(router *server.Router) {
	router.GET("/v2/healthcheck/ping", handlePing)
}

func handlePing(req *http.Request, h *server.HandlerToolkit) (*server.Response, error) {
	return h.Response(server.JSON{"message": "pong"}), nil
}
