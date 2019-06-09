package transactions

import (
	"net/http"
	"strings"
	"time"

	"ledger.api/pkg/auth"

	"ledger.api/pkg/core/router"
)

func createSummaryQueryHandler(svc QueryService) router.ToolkitHandlerFunc {
	handleTimeValue := func(rawValue string) (interface{}, error) {
		if rawValue == "" {
			return (*time.Time)(nil), nil
		}
		from, err := time.Parse(time.RFC3339, rawValue)
		if err != nil {
			return nil, err
		}
		return &from, nil
	}

	handleCSVValue := func(rawValue string) (interface{}, error) {
		values := strings.FieldsFunc(rawValue, func(c rune) bool {
			return c == ','
		})
		if len(values) == 0 {
			return ([]string)(nil), nil
		}
		return values, nil
	}

	return func(w http.ResponseWriter, r *http.Request, h router.HandlerToolkit) error {
		var params struct {
			LedgerID      string
			Type          string
			From          *time.Time
			To            *time.Time
			ExcludeTagIDs []string
		}

		err := h.BindParams().
			PathParam("ledgerID").String(&params.LedgerID).
			PathParam("type").String(&params.Type).
			QueryParam("from").Custom(&params.From, handleTimeValue).
			QueryParam("to").Custom(&params.To, handleTimeValue).
			QueryParam("excludeTagIDs").Custom(&params.ExcludeTagIDs, handleCSVValue).
			Validate(&params)
		if err != nil {
			return err
		}
		query := newSummaryQuery(params.LedgerID, params.Type,
			optionalDates(params.From, params.To),
			withExcludeTagIDs(params.ExcludeTagIDs),
		)
		result, err := svc.processSummaryQuery(r.Context(), query)
		if err != nil {
			return err
		}
		return h.WriteJSON(result)
	}
}

// SetupRoutes will register transactions routes
func SetupRoutes(appRouter router.Router, svc QueryService) {
	// TODO: some authorization if user is authorized for a requested ledger

	// from=:from&to=:to&excludeTags=:excludeTagIDs
	appRouter.Handle("GET",
		"/v2/ledgers/:ledgerID/transactions/:type/summary",
		auth.AuthorizeRequest(createSummaryQueryHandler(svc), auth.AllowScope("read:transactions")),
	)
}
