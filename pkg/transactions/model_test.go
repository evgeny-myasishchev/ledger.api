package transactions

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"ledger.api/pkg/internal/ledgertesting"
	"ledger.api/pkg/tags"
)

func TestProcessSummaryQuery(t *testing.T) {
	svc := createQueryService(DB)
	ctx := context.Background()

	FocusConvey("Given summaryQuery", t, func() {
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

		FocusConvey("When type is expense", func() {
			rndTag := ledgertesting.TrxRndTag(md.TagIDs)
			trxDate := ledgertesting.TrxDate(time.Now())
			var trxs [100]ledgertesting.Transaction
			for i := 0; i < 100; i++ {
				trxs[i] = *ledgertesting.NewExpenseTransaction(rndTag, trxDate)
			}
			err := ledgertesting.SetupTransactions(DB, trxs[:])
			So(err, ShouldBeNil)
			query := summaryQuery{ledgerID: md.LedgerID, typ: "expense"}

			FocusConvey("It should calculate summary for the past month grouped by tag", func() {
				expectedByTagID := make(map[int]*summaryDTO)
				expectedResults := []summaryDTO{}
				for _, trx := range trxs {
					tagID := tags.GetTagIDsFromString(trx.TagIDs)[0]
					if expectedByTagID[tagID] == nil {
						expectedByTagID[tagID] = &summaryDTO{
							tagID:   tagID,
							tagName: md.TagsByID[tagID],
							amount:  trx.Amount,
						}
					} else {
						expectedByTagID[tagID].amount += trx.Amount
					}
				}
				for _, v := range expectedByTagID {
					expectedResults = append(expectedResults, *v)
				}

				sort.Slice(expectedResults, func(li, ri int) bool {
					return expectedResults[li].amount > expectedResults[ri].amount
				})
				result, err := svc.processSummaryQuery(ctx, &query)
				So(err, ShouldBeNil)
				So(result, ShouldResemble, expectedResults)
			})

			Convey("It should not count refunds", func() {
			})

			Convey("It should not include transactions from other ledgers", func() {
			})

			Convey("It should exclude provided tags", func() {
			})

			Convey("It should exclude transactions outside of date range", func() {
			})
		})
	})
}
