package transactions

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/icrowley/fake"
	"github.com/satori/go.uuid"

	"ledger.api/pkg/logging"

	. "github.com/smartystreets/goconvey/convey"
	"ledger.api/pkg/internal/ledgertesting"
	"ledger.api/pkg/tags"
)

func TestNewSummaryQuery(t *testing.T) {
	Convey("It should create a new instance of the query with defaults", t, func() {
		typ := fake.Word()
		ledgerID := uuid.NewV4().String()
		query := newSummaryQuery(ledgerID, typ)
		So(query.ledgerID, ShouldEqual, ledgerID)
		So(query.typ, ShouldEqual, typ)
		now := time.Now()
		So(query.from, ShouldNotBeNil)
		So(query.from.Unix(), ShouldAlmostEqual, now.AddDate(0, -1, 0).Unix())

		So(query.to, ShouldNotBeNil)
		So(query.to.Unix(), ShouldAlmostEqual, now.Unix())
	})
}

func TestProcessSummaryQuery(t *testing.T) {
	svc := CreateQueryService(DB)
	ctx := logging.CreateContext(context.Background(), logging.NewTestLogger())

	Convey("Given summaryQuery", t, func() {
		md, err := ledgertesting.SetupLedgerData(DB)
		So(err, ShouldBeNil)

		Convey("When required parameters are missing", func() {
			Convey("It should return error if no ledger provided", func() {
				_, err := svc.processSummaryQuery(ctx, &summaryQuery{typ: "income"})
				So(err, ShouldResemble, errors.New("Please provide ledgerID"))
			})

			Convey("It should return error if no type provided", func() {
				_, err := svc.processSummaryQuery(ctx, &summaryQuery{ledgerID: "nil-ledger"})
				So(err, ShouldResemble, errors.New("Please provide type"))
			})
		})

		Convey("When type is expense", func() {
			rndTag := ledgertesting.TrxRndTag(md.TagIDs)
			rndAcc := ledgertesting.TrxRndAcc(md.AccountIDs)
			dateMax := time.Now()
			dateMin := dateMax.AddDate(0, -1, 0)
			trxDate := ledgertesting.TrxRndDate(dateMin, dateMax)
			var trxs [100]ledgertesting.Transaction
			for i := 0; i < 100; i++ {
				trxs[i] = *ledgertesting.NewExpenseTransaction(rndTag, trxDate, rndAcc)
			}
			err := ledgertesting.SetupTransactions(DB, trxs[:])
			So(err, ShouldBeNil)
			query := summaryQuery{ledgerID: md.LedgerID, typ: "expense"}

			Convey("It should calculate summary grouped by tag", func() {
				expectedByTagID := make(map[int]*summaryDTO)
				expectedResults := []summaryDTO{}
				for _, trx := range trxs {
					tagID := tags.GetTagIDsFromString(trx.TagIDs)[0]
					if expectedByTagID[tagID] == nil {
						expectedByTagID[tagID] = &summaryDTO{
							TagID:   tagID,
							TagName: md.TagsByID[tagID],
							Amount:  trx.Amount,
						}
					} else {
						expectedByTagID[tagID].Amount += trx.Amount
					}
				}
				for _, v := range expectedByTagID {
					expectedResults = append(expectedResults, *v)
				}

				sort.Slice(expectedResults, func(li, ri int) bool {
					return expectedResults[li].Amount > expectedResults[ri].Amount
				})
				actualResult, err := svc.processSummaryQuery(ctx, &query)
				So(err, ShouldBeNil)
				So(len(actualResult), ShouldEqual, len(expectedResults))
				for i, actualSummary := range actualResult {
					expectedSummary := expectedResults[i]
					So(expectedSummary, ShouldResemble, actualSummary)
				}
			})

			Convey("It should not include zero tags", func() {
				var maxTagID int
				for _, tagID := range md.TagIDs {
					if maxTagID > tagID {
						maxTagID = tagID
					}
				}

				zeroTagID1 := maxTagID + rnd.Intn(2000)
				zeroTagID2 := maxTagID + rnd.Intn(2000)
				if err := ledgertesting.SetupTag(DB, md.LedgerID, zeroTagID1, "zero tag 1"); err != nil {
					So(err, ShouldBeNil)
				}
				if err := ledgertesting.SetupTag(DB, md.LedgerID, zeroTagID2, "zero tag 2"); err != nil {
					So(err, ShouldBeNil)
				}

				result, err := svc.processSummaryQuery(ctx, &query)
				So(err, ShouldBeNil)

				for _, summary := range result {
					So(summary.TagID, ShouldNotEqual, zeroTagID1)
					So(summary.TagID, ShouldNotEqual, zeroTagID2)
				}
			})

			Convey("It should exclude transactions outside of date range", func() {
				var maxTagID int
				for _, tagID := range md.TagIDs {
					if maxTagID > tagID {
						maxTagID = tagID
					}
				}
				tagID1 := maxTagID + rnd.Intn(2000)
				tagID2 := maxTagID + rnd.Intn(2000)
				if err := ledgertesting.SetupTag(DB, md.LedgerID, tagID1, "tag outside 1"); err != nil {
					So(err, ShouldBeNil)
				}
				if err := ledgertesting.SetupTag(DB, md.LedgerID, tagID2, "tag outside 2"); err != nil {
					So(err, ShouldBeNil)
				}

				rndTagsOutside := ledgertesting.TrxRndTag([]int{tagID1, tagID2})

				trx1 := ledgertesting.NewExpenseTransaction(
					rndTagsOutside,
					ledgertesting.TrxRndDate(dateMin.AddDate(0, -2, 0), dateMin.AddDate(0, 0, -1)),
					rndAcc,
				)
				trx2 := ledgertesting.NewExpenseTransaction(
					rndTagsOutside,
					ledgertesting.TrxRndDate(dateMax.AddDate(0, 0, 1), dateMax.AddDate(0, 1, 0)),
					rndAcc,
				)
				err := ledgertesting.SetupTransactions(DB, []ledgertesting.Transaction{
					*trx1,
					*trx2,
				})
				So(err, ShouldBeNil)

				query.from = &dateMin
				query.to = &dateMax
				actualResult, err := svc.processSummaryQuery(ctx, &query)
				So(err, ShouldBeNil)
				for _, summary := range actualResult {
					So(summary.TagID, ShouldNotEqual, tagID1)
					So(summary.TagID, ShouldNotEqual, tagID2)
				}
			})

			Convey("It should subtract refunds from expense", func() {
			})

			Convey("It should not include transactions from other ledgers", func() {
			})

			Convey("It should exclude provided tags", func() {
			})
		})

		Convey("When account currency is other than ledger default", func() {
			Convey("It should convert the amount to default currency prior to calculation", func() {

			})
		})
	})
}
