package ledgers

import (
	"os"
	"testing"

	"github.com/icrowley/fake"
	"github.com/jinzhu/gorm"
	. "github.com/smartystreets/goconvey/convey"
	"ledger.api/pkg/app"
)

var DB *gorm.DB
var service Service

func TestMain(m *testing.M) {
	cfg := app.GetConfig()
	DB = app.OpenGormConnection(cfg.GetString("DB_URL"))
	defer DB.Close()
	service = CreateService(DB)

	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}

func randomLedger() *Ledger {
	l := Ledger{
		Name:          fake.Brand(),
		CreatedUserID: fake.Characters(),
	}
	return &l
}

func TestLedgerService(t *testing.T) {
	Convey("Given new ledger object", t, func() {
		newLedger := randomLedger()
		Convey("When created", func() {
			created, err := service.createLedger(*newLedger)

			Convey("It should not fail", func() {
				So(err, ShouldBeNil)
			})

			Convey("It should get ID generated", func() {
				So(created.ID, ShouldNotBeEmpty)
			})

			Convey("It should be written to the db", func() {
				fromDb := Ledger{}
				res := DB.Where("id = ?", created.ID).First(&fromDb)
				fromDb.CreatedAt = created.CreatedAt
				fromDb.UpdatedAt = created.UpdatedAt
				So(res.RecordNotFound(), ShouldBeFalse)
				So(&fromDb, ShouldResemble, created)
			})
		})
	})
}
