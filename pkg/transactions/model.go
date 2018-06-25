package transactions

import (
	"context"
	"errors"
	"time"

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
	return nil, errors.New("Not implemented")
}

func createQueryService(db *gorm.DB) queryService {
	svc := dbQueryService{db: db}
	return &svc
}
