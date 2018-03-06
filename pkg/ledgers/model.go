package ledgers

import "time"

// Ledger model
type Ledger struct {
	ID            string    `gorm:"primary_key;type:character varying NOT NULL"`
	Name          string    `gorm:"type:character varying NOT NULL"`
	CreatedUserID string    `gorm:"type:character varying NOT NULL"`
	CurrencyCode  string    `gorm:"type:character varying NOT NULL"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`
}
