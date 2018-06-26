package transactions

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"ledger.api/pkg/internal/ledgertesting"
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
			FocusConvey("It should calculate summary for the past month grouped by tag", func() {
				rndTag := ledgertesting.TrxRndTag(md.TagIDs)
				trxDate := ledgertesting.TrxDate(time.Now())
				trxs := []ledgertesting.Transaction{
					*ledgertesting.NewExpenseTransaction(rndTag, trxDate),
					*ledgertesting.NewExpenseTransaction(rndTag, trxDate),
					*ledgertesting.NewExpenseTransaction(rndTag, trxDate),
					*ledgertesting.NewExpenseTransaction(rndTag, trxDate),
				}
				err := ledgertesting.SetupTransactions(DB, trxs)
				So(err, ShouldBeNil)

				println("==== hello =====")
				fmt.Printf("%+v\n", md)
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
