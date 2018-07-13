package ledgers

import (
	"context"

	"github.com/jinzhu/gorm"
)

type ledgerDTO struct {
	LedgerID     string `json:"ledgerID" gorm:"column:aggregate_id"`
	Name         string `json:"name"`
	CurrencyCode string `json:"currencyCode"`
}

type userLedgersQuery struct {
}

// QueryService is a service to do various queries against ledgers
type QueryService interface {
	processUserLedgersQuery(ctx context.Context, query *userLedgersQuery) ([]ledgerDTO, error)
}

type dbQueryService struct {
	db *gorm.DB
}

func (svc *dbQueryService) processUserLedgersQuery(ctx context.Context, query *userLedgersQuery) ([]ledgerDTO, error) {
	result := []ledgerDTO{}
	if err := svc.db.Table("projections_ledgers ldr").
		Select("ldr.aggregate_id, ldr.name, ldr.currency_code").
		Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

// CreateQueryService initializes a new instance of the query service
func CreateQueryService(db *gorm.DB) QueryService {
	svc := dbQueryService{db: db}
	return &svc
}
