package transactions

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"ledger.api/pkg/logging"

	. "github.com/smartystreets/goconvey/convey"
	"ledger.api/pkg/internal/ledgertesting"
	"ledger.api/pkg/tags"
)

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
			trxDate := ledgertesting.TrxDate(time.Now())
			var trxs [100]ledgertesting.Transaction
			for i := 0; i < 100; i++ {
				trxs[i] = *ledgertesting.NewExpenseTransaction(rndTag, trxDate, rndAcc)
			}
			err := ledgertesting.SetupTransactions(DB, trxs[:])
			So(err, ShouldBeNil)
			query := summaryQuery{ledgerID: md.LedgerID, typ: "expense"}

			Convey("It should calculate summary for the past month grouped by tag", func() {
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

			Convey("It should subtract refunds from expense", func() {
			})

			Convey("It should not include transactions from other ledgers", func() {
			})

			Convey("It should exclude provided tags", func() {
			})

			Convey("It should exclude transactions outside of date range", func() {
			})
		})

		Convey("When account currency is other than ledger default", func() {
			Convey("It should convert the amount to default currency prior to calculation", func() {

			})
		})
	})
}
