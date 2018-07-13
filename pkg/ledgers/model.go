package ledgers

import "context"

type ledgerDTO struct {
	ledgerID     string `json:"ledgerID"`
	name         string `json:"name"`
	currencyCode string `json:"currencyCode"`
}

type userLedgersQuery struct {
}

// QueryService is a service to do various queries against ledgers
type QueryService interface {
	processUserLedgersQuery(ctx context.Context, query *userLedgersQuery) ([]ledgerDTO, error)
}
