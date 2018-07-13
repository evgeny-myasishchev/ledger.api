package ledgers

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/icrowley/fake"
	"github.com/satori/go.uuid"

	"github.com/jinzhu/gorm"
	. "github.com/smartystreets/goconvey/convey"
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

type ledger struct {
	id                int `gorm:"primary_key"`
	aggregateID       string
	ownerUserID       int
	name              string
	currencyCode      string
	authorizedUserIDs string
}

func (ledger) TableName() string {
	return "projections_ledgers"
}

func setupLedgers(db *gorm.DB) ([]ledger, error) {
	numLedgers := 2 + rnd.Intn(5)
	ledgers := make([]ledger, numLedgers)
	fmt.Println("fuck")
	for i := 0; i < numLedgers; i++ {
		ldr := ledger{
			aggregateID:  uuid.NewV4().String(),
			ownerUserID:  rnd.Intn(20),
			name:         fake.Brand(),
			currencyCode: fake.CurrencyCode(),
		}
		ledgers[i] = ldr
		if err := db.Create(&ldr).Error; err != nil {
			fmt.Println("Returning err")
			return nil, err
		}
	}
	fmt.Println("Returning")
	return ledgers, errors.New("Not implemented")
}

func TestUserLedgersQuery(t *testing.T) {
	Convey("Given user ledgers query", t, func() {
		ledgers, err := setupLedgers(DB)
		So(err, ShouldBeNil)
		Convey("When default query object is used", func() {
			Convey("It should return all ledgers", func() {
				fmt.Printf("ledgers %+v", ledgers)
			})
		})
	})
}
