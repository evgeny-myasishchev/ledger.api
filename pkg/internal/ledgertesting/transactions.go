package ledgertesting

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/icrowley/fake"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

// LedgerData represents ledger related data used in tests
type LedgerData struct {
	ledgerID   string
	TagIDs     []int
	TagsByID   map[int]string
	TagsByName map[string]int
}

// Transaction represents test transaction related data to setup for tests
type Transaction struct {
	typeID     int
	Amount     int
	tagIDs     string
	comment    string
	date       time.Time
	isTransfer bool
}

// TransactionSetup fn to setup transaction
type TransactionSetup func(*Transaction)

// TrxDate setup transaction for given date
func TrxDate(date time.Time) TransactionSetup {
	return func(trx *Transaction) {
		trx.date = date
	}
}

// RndTag will assign a random tag for given trx
func TrxRndTag(tagIDs []int) TransactionSetup {
	return func(trx *Transaction) {
		trx.tagIDs = fmt.Sprintf("{%v}", tagIDs[rnd.Intn(len(tagIDs))])
	}
}

// NewExpenseTransaction creates a new mock transaction structure
func NewExpenseTransaction(setup ...TransactionSetup) *Transaction {
	trx := Transaction{
		typeID:  2,
		Amount:  1000 + rnd.Intn(100000),
		date:    time.Now(),
		comment: fmt.Sprintf("Test income trx %v", fake.Word()),
	}
	for _, setupFn := range setup {
		setupFn(&trx)
	}
	return &trx
}

// SetupLedgerData generate and persist mock ledger data
func SetupLedgerData(db *gorm.DB) (LedgerData, error) {
	ledgerID := uuid.NewV4().String()
	md := LedgerData{
		ledgerID:   ledgerID,
		TagIDs:     make([]int, 10),
		TagsByID:   make(map[int]string),
		TagsByName: make(map[string]int),
	}

	base := rnd.Intn(10000)
	for i := 0; i < 10; i++ {
		tagID := i + base
		tagName := fmt.Sprintf("Tag %v %v", base, i)
		md.TagIDs[i] = tagID
		md.TagsByID[tagID] = tagName
		md.TagsByName[tagName] = tagID

		if err := db.Exec(`
			INSERT INTO projections_tags(ledger_id, tag_id, name, authorized_user_ids)
			VALUES(?,?,?,'')
			`, ledgerID, tagID, tagName).Error; err != nil {
			return LedgerData{}, err
		}
	}

	return md, nil
}

// SetupTransactions persist mock transactions to db
func SetupTransactions(db *gorm.DB, transactions []Transaction) error {
	for _, trx := range transactions {
		transactionID := uuid.NewV4().String()
		accountID := uuid.NewV4().String()
		if err := db.Debug().Exec(`
			INSERT INTO projections_transactions(
				transaction_id,
				account_id,
				type_id,
				amount,
				tag_ids,
				comment,
				date,
				is_transfer
			)
			VALUES(?,?,?,?,?,?,?,?)
			`, transactionID, accountID, trx.typeID, trx.Amount, trx.tagIDs, trx.comment, trx.date, trx.isTransfer,
		).Error; err != nil {
			return err
		}
	}
	return nil
}
