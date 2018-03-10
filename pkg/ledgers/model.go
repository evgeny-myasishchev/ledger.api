package ledgers

import (
	"time"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

// Ledger model
type Ledger struct {
	ID            string    `gorm:"primary_key;type:character varying NOT NULL"`
	Name          string    `gorm:"type:character varying NOT NULL"`
	CreatedUserID string    `gorm:"type:character varying NOT NULL"`
	CurrencyCode  string    `gorm:"type:character varying NOT NULL"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`
}

// Service that defines common ledger operations operations
type Service interface {
	createLedger(Ledger) (*Ledger, error)
}

type dbService struct {
	db *gorm.DB
}

func (dbSvc dbService) createLedger(ledger Ledger) (*Ledger, error) {
	ledgerID := uuid.NewV4().String()
	ledger.ID = ledgerID
	dbSvc.db.Create(&ledger)
	return &ledger, nil
}

// CreateService - creates ledger service implementation
func CreateService(db *gorm.DB) Service {
	svc := dbService{db: db}
	return &svc
}

// ResetSchema - recreate db tables
func ResetSchema(db *gorm.DB) {
	db.DropTable(&Ledger{})
	db.CreateTable(&Ledger{})
}
