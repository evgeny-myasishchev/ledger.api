package transactions

import (
	"context"
	"errors"
	"time"

	"ledger.api/pkg/logging"

	"github.com/jinzhu/gorm"
)

type summaryDTO struct {
	TagID   int    `json:"tagID"`
	TagName string `json:"tagName"`
	Amount  int    `json:"amount"`
}

type summaryQuery struct {
	ledgerID      string
	typ           string
	from          *time.Time
	to            *time.Time
	excludeTagIDs []string
}

func newSummaryQuery(ledgerID string, typ string, queryInit ...func(*summaryQuery)) *summaryQuery {
	now := time.Now()
	from := now.AddDate(0, -1, 0)
	query := &summaryQuery{
		ledgerID: ledgerID,
		typ:      typ,
		from:     &from,
		to:       &now,
	}
	for _, initFn := range queryInit {
		initFn(query)
	}
	return query
}

// QueryService is a service to do various gueries against transactions
type QueryService interface {
	processSummaryQuery(ctx context.Context, query *summaryQuery) ([]summaryDTO, error)
}

type dbQueryService struct {
	db *gorm.DB
}

func (svc *dbQueryService) processSummaryQuery(ctx context.Context, query *summaryQuery) ([]summaryDTO, error) {
	if query.ledgerID == "" {
		return nil, errors.New("Please provide ledgerID")
	}
	if query.typ == "" {
		return nil, errors.New("Please provide type")
	}
	logger := logging.FromContext(ctx)
	logger.Debugf("Processing summary query. LedgerID: %v", query.ledgerID)
	result := []summaryDTO{}

	from := query.from
	if from == nil {
		from = &time.Time{}
	}
	to := query.to
	if to == nil {
		now := time.Now()
		to = &now
	}

	rows, err := svc.db.Raw(`
		SELECT tg.tag_id tagID, tg.name tagName, SUM(trx.amount) amount
		FROM projections_transactions trx
			JOIN projections_accounts acc 
				ON acc.aggregate_id = trx.account_id
			JOIN projections_tags tg 
				ON tg.ledger_id = acc.ledger_id AND trx.tag_ids LIKE '%{'||tg.tag_id||'}%'
		WHERE acc.ledger_id = ?
			AND trx.date >= ? 
			AND trx.date <= ?
		GROUP BY tg.tag_id, tg.name
		ORDER BY amount DESC`, query.ledgerID, from, to).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var tagID, amount int
		var tagName string
		if err = rows.Scan(&tagID, &tagName, &amount); err != nil {
			return nil, err
		}
		result = append(result, summaryDTO{TagID: tagID, TagName: tagName, Amount: amount})
	}
	return result, nil
}

// CreateQueryService initializes a new instance of the query service
func CreateQueryService(db *gorm.DB) QueryService {
	svc := dbQueryService{db: db}
	return &svc
}
