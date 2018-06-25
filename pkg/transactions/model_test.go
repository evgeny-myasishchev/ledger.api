package transactions

import (
	"context"
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestProcessSummaryQuery(t *testing.T) {
	svc := createQueryService(DB)
	ctx := context.Background()

	Convey("Given summaryQuery", t, func() {
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
			Convey("It should calculate summary for the past month grouped by tag", func() {
			})

			Convey("It should not count refunds", func() {
			})

			Convey("It should not include transactions from other ledgers", func() {
			})

			Convey("It should exclude provided tags", func() {
			})
		})
	})
}
