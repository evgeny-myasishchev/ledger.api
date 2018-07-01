package ldtesting

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
	LedgerID   string
	TagIDs     []int
	TagsByID   map[int]string
	TagsByName map[string]int
	AccountIDs []string
}

// Transaction represents test transaction related data to setup for tests
type Transaction struct {
	typeID     int
	Amount     int
	AccountID  string
	TagIDs     string
	comment    string
	Date       time.Time
	isTransfer bool
}

// TransactionSetup fn to setup transaction
type TransactionSetup func(*Transaction)

// TrxDate setup transaction for given date
func TrxDate(date time.Time) TransactionSetup {
	return func(trx *Transaction) {
		trx.Date = date
	}
}

// TrxRndDate setup transaction with random date from given range
func TrxRndDate(min time.Time, max time.Time) TransactionSetup {
	diff := max.Unix() - min.Unix()
	return func(trx *Transaction) {
		trx.Date = time.Unix(min.Unix()+rnd.Int63n(diff), 0)
	}
}

// TrxRndTag will assign a random tag for given trx
func TrxRndTag(tagIDs []int) TransactionSetup {
	return func(trx *Transaction) {
		trx.TagIDs = fmt.Sprintf("{%v}", tagIDs[rnd.Intn(len(tagIDs))])
	}
}

// TrxRndAcc will assign a random account for given trx
func TrxRndAcc(accountIDs []string) TransactionSetup {
	return func(trx *Transaction) {
		trx.AccountID = accountIDs[rnd.Intn(len(accountIDs))]
	}
}

// NewExpenseTransaction creates a new mock transaction structure
func NewExpenseTransaction(setup ...TransactionSetup) *Transaction {
	trx := Transaction{
		typeID:  2,
		Amount:  1000 + rnd.Intn(100000),
		Date:    time.Now(),
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
		LedgerID:   ledgerID,
		TagIDs:     make([]int, 10),
		TagsByID:   make(map[int]string),
		TagsByName: make(map[string]int),
		AccountIDs: make([]string, 10),
	}

	base := rnd.Intn(10000)
	for i := 0; i < 10; i++ {
		tagID := i + base
		tagName := fmt.Sprintf("Tag %v %v", base, i)
		md.TagIDs[i] = tagID
		md.TagsByID[tagID] = tagName
		md.TagsByName[tagName] = tagID

		if err := SetupTag(db, ledgerID, tagID, tagName); err != nil {
			return LedgerData{}, err
		}

		accountID := uuid.NewV4().String()
		md.AccountIDs[i] = accountID

		if err := db.Exec(`
			INSERT INTO projections_accounts(
				ledger_id,
				aggregate_id,
				sequential_number,
				owner_user_id,
				authorized_user_ids,
				currency_code,
				name,
				balance,
				is_closed
			)
			VALUES(?,?, ?, 0, '', 'UAH', ?, 0, false)
			`, ledgerID, accountID, i, fmt.Sprintf("Account %v-%v", base, i)).Error; err != nil {
			return LedgerData{}, err
		}
	}

	return md, nil
}

// SetupTag will insert a new tag into tags projection
func SetupTag(db *gorm.DB, ledgerID string, tagID int, tagName string) error {
	return db.Exec(`
		INSERT INTO projections_tags(ledger_id, tag_id, name, authorized_user_ids)
		VALUES(?,?,?,'')
		`, ledgerID, tagID, tagName).Error
}

// SetupTransactions persist mock transactions to db
func SetupTransactions(db *gorm.DB, transactions []Transaction) error {
	for _, trx := range transactions {
		transactionID := uuid.NewV4().String()
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
			`, transactionID, trx.AccountID, trx.typeID, trx.Amount, trx.TagIDs, trx.comment, trx.Date, trx.isTransfer,
		).Error; err != nil {
			return err
		}
	}
	return nil
}
