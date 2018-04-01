package ledgers

import (
	"fmt"
	"testing"

	"github.com/icrowley/fake"
	uuid "github.com/satori/go.uuid"
	. "github.com/smartystreets/goconvey/convey"
	"ledger.api/pkg/users"
)

func randomUser() *users.User {
	u := users.User{
		ID: fmt.Sprintf("user-%v", fake.Word()),
	}
	return &u
}

func randomNewLedger() *NewLedger {
	l := NewLedger{
		ID:   uuid.NewV4().String(),
		Name: fake.Brand(),
	}
	return &l
}

func TestLedgerService(t *testing.T) {
	Convey("Given new ledger object", t, func() {
		user := randomUser()
		newLedger := randomNewLedger()
		Convey("When created", func() {
			created, err := service.createLedger(user, newLedger)

			Convey("It should not fail", func() {
				So(err, ShouldBeNil)
			})

			Convey("Create a new ledger for given attributes", func() {
				So(created.ID, ShouldEqual, newLedger.ID)
				So(created.Name, ShouldEqual, newLedger.Name)
				So(created.CreatedUserID, ShouldEqual, user.ID)
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

		Convey("When failed to create", func() {
			created, err := service.createLedger(user, &NewLedger{})
			Convey("It should return error object", func() {
				So(err, ShouldNotBeNil)
				So(created, ShouldBeNil)
			})
		})
	})
}
