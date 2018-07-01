package transactions

import (
	"net/http"
	"strings"
	"time"

	"ledger.api/pkg/server"
)

// CreateRoutes - Register transactions related routes
func CreateRoutes(svc QueryService) server.Routes {
	return func(router *server.Router) {
		router.GET(
			// from=:from&to=:to&excludeTags=:excludeTagIDs
			"/v2/ledgers/:ledgerID/transactions/:type/summary",
			// server.RequireScopes(processSummaryQuery, "read:transactions"),
			createSummaryQueryHandler(svc),
		)
	}
}

func parseQueryTime(req *http.Request, key string) (*time.Time, error) {
	if timeStr := req.URL.Query().Get(key); timeStr != "" {
		from, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			return nil, err
		}
		return &from, nil
	}
	return nil, nil
}

func createSummaryQueryHandler(svc QueryService) server.HandlerFunc {
	return func(req *http.Request, h *server.HandlerToolkit) (*server.Response, error) {
		ledgerID := h.Params.ByName("ledgerID")
		typ := h.Params.ByName("type")
		from, err := parseQueryTime(req, "from")
		if err != nil {
			return nil, err
		}
		to, err := parseQueryTime(req, "to")
		if err != nil {
			return nil, err
		}
		query := newSummaryQuery(ledgerID, typ, func(q *summaryQuery) {
			q.from = from
			q.to = to
		})
		if val := req.URL.Query().Get("excludeTagIDs"); val != "" {
			query.excludeTagIDs = strings.Split(val, ",")
		}
		result, err := svc.processSummaryQuery(req.Context(), query)
		if err != nil {
			return nil, err
		}
		return h.Response(result), nil
	}
}
