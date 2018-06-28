package transactions

import (
	"context"
	"errors"
	"time"

	"ledger.api/pkg/logging"

	"github.com/jinzhu/gorm"
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
	rows, err := svc.db.Raw(`
		SELECT tg.tag_id tagID, tg.name tagName, SUM(trx.amount) amount
		FROM projections_transactions trx
			JOIN projections_accounts acc ON acc.aggregate_id = trx.account_id
			JOIN projections_tags tg ON trx.tag_ids LIKE '%{'||tg.tag_id||'}%'
		WHERE acc.ledger_id = ?
		GROUP BY tg.tag_id, tg.name
		ORDER BY amount DESC`, query.ledgerID).Rows()
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
		result = append(result, summaryDTO{tagID: tagID, tagName: tagName, amount: amount})
	}
	return result, nil
}

func createQueryService(db *gorm.DB) queryService {
	svc := dbQueryService{db: db}
	return &svc
}
