package ledgers

import (
	"time"

	"github.com/jinzhu/gorm"
	"ledger.api/pkg/users"
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

// NewLedger model
type NewLedger struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Service that defines common ledger operations operations
type Service interface {
	createLedger(user *users.User, newLedger *NewLedger) (*Ledger, error)
}

type dbService struct {
	db *gorm.DB
}

func (dbSvc dbService) createLedger(user *users.User, newLedger *NewLedger) (*Ledger, error) {
	ledger := Ledger{
		ID:            newLedger.ID,
		Name:          newLedger.Name,
		CreatedUserID: user.ID,
	}
	if err := dbSvc.db.Create(&ledger).Error; err != nil {
		return nil, err
	}
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
