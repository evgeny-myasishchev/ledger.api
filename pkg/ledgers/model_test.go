package ledgers

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/icrowley/fake"
	"github.com/satori/go.uuid"
	funk "github.com/thoas/go-funk"
	"ledger.api/pkg/logging"

	"github.com/jinzhu/gorm"
	. "github.com/smartystreets/goconvey/convey"
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

type ledger struct {
	ID                int `gorm:"primary_key"`
	AggregateID       string
	OwnerUserID       int
	Name              string
	CurrencyCode      string
	AuthorizedUserIDs string
}

func (ledger) TableName() string {
	return "projections_ledgers"
}

func setupLedgers(db *gorm.DB) ([]ledger, error) {
	if err := db.Delete(&ledger{}).Error; err != nil {
		return nil, err
	}

	numLedgers := 2 + rnd.Intn(5)
	ledgers := make([]ledger, numLedgers)
	for i := 0; i < numLedgers; i++ {
		ldr := ledger{
			AggregateID:  uuid.NewV4().String(),
			OwnerUserID:  rnd.Intn(20),
			Name:         fake.Brand(),
			CurrencyCode: fake.CurrencyCode(),
		}
		ledgers[i] = ldr
		if err := db.Create(&ldr).Error; err != nil {
			return nil, err
		}
	}
	return ledgers, nil
}

func TestUserLedgersQuery(t *testing.T) {
	Convey("Given user ledgers query", t, func() {
		ctx := logging.CreateContext(context.Background(), logging.NewTestLogger())
		ledgers, err := setupLedgers(DB)
		svc := CreateQueryService(DB)
		So(err, ShouldBeNil)
		Convey("When default query object is used", func() {
			Convey("It should return all ledgers", func() {
				ledgersDTOs, err := svc.processUserLedgersQuery(ctx, &userLedgersQuery{})
				So(err, ShouldBeNil)
				So(ledgersDTOs, ShouldHaveLength, len(ledgers))
				ledgersByID := funk.ToMap(ledgers, "AggregateID").(map[string]ledger)
				for _, dto := range ledgersDTOs {
					actualLedger := ledgersByID[dto.LedgerID]
					So(dto, ShouldResemble, ledgerDTO{
						LedgerID:     actualLedger.AggregateID,
						Name:         actualLedger.Name,
						CurrencyCode: actualLedger.CurrencyCode,
					})
				}
			})
		})
	})
}
