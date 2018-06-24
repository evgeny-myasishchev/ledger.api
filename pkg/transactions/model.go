package transactions

import (
	"context"
	"time"
)

type summaryDTO struct {
	tagID   int
	tagName string
	amount  int
}

type summaryQuery struct {
	ledgerID    string
	typ         string
	from        *time.Time
	to          *time.Time
	excludeTags []int
}

type queryService interface {
	processSummaryQuery(ctx context.Context, query *summaryQuery) ([]summaryDTO, error)
}
