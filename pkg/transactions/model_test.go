package transactions

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/icrowley/fake"
	"github.com/satori/go.uuid"

	. "github.com/smartystreets/goconvey/convey"
	"ledger.api/pkg/internal/ldtesting"
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
		md, err := ldtesting.SetupLedgerData(DB)
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
			rndTag := ldtesting.TrxRndTag(md.TagIDs)
			rndAcc := ldtesting.TrxRndAcc(md.AccountIDs)
			dateMax := time.Now()
			dateMin := dateMax.AddDate(0, -1, 0)
			trxDate := ldtesting.TrxRndDate(dateMin, dateMax)
			var trxs [100]ldtesting.Transaction
			for i := 0; i < 100; i++ {
				trxs[i] = *ldtesting.NewTransaction(rndTag, trxDate, rndAcc)
				// TODO: Add some income as well to make sure it's excluded
			}
			err := ldtesting.SetupTransactions(DB, trxs[:])
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

			Convey("It not include income", func() {
				expectedByTagID := make(map[int]*summaryDTO)
				expectedResults := []summaryDTO{}

				rndTagIncome := ldtesting.TrxRndTag(md.TagIDs)
				trxin1 := ldtesting.NewTransaction(rndTagIncome, trxDate, rndAcc, ldtesting.TrxIncome)
				trxin2 := ldtesting.NewTransaction(rndTagIncome, trxDate, rndAcc, ldtesting.TrxIncome)
				ldtesting.SetupTransactions(DB, []ldtesting.Transaction{*trxin1, *trxin2})

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
				if err := ldtesting.SetupTag(DB, md.LedgerID, zeroTagID1, "zero tag 1"); err != nil {
					So(err, ShouldBeNil)
				}
				if err := ldtesting.SetupTag(DB, md.LedgerID, zeroTagID2, "zero tag 2"); err != nil {
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
				if err := ldtesting.SetupTag(DB, md.LedgerID, tagID1, "tag outside 1"); err != nil {
					So(err, ShouldBeNil)
				}
				if err := ldtesting.SetupTag(DB, md.LedgerID, tagID2, "tag outside 2"); err != nil {
					So(err, ShouldBeNil)
				}

				rndTagsOutside := ldtesting.TrxRndTag([]int{tagID1, tagID2})

				trx1 := ldtesting.NewTransaction(
					rndTagsOutside,
					ldtesting.TrxRndDate(dateMin.AddDate(0, -2, 0), dateMin.AddDate(0, 0, -1)),
					rndAcc,
				)
				trx2 := ldtesting.NewTransaction(
					rndTagsOutside,
					ldtesting.TrxRndDate(dateMax.AddDate(0, 0, 1), dateMax.AddDate(0, 1, 0)),
					rndAcc,
				)
				err := ldtesting.SetupTransactions(DB, []ldtesting.Transaction{
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

			Convey("It should exclude provided tags", func() {
				expectedByTagID := make(map[int]*summaryDTO)
				expectedResults := []summaryDTO{}
				tagsToExcludeMap := make(map[int]bool)
				tagsToExclude := []string{}
				for len(tagsToExclude) <= 2 {
					trx := trxs[rnd.Intn(len(trxs))]
					tagID := tags.GetTagIDsFromString(trx.TagIDs)[0]
					if _, ok := tagsToExcludeMap[tagID]; !ok {
						tagsToExcludeMap[tagID] = true
						tagsToExclude = append(tagsToExclude, strconv.Itoa(tagID))
					}
				}

				for _, trx := range trxs {
					tagID := tags.GetTagIDsFromString(trx.TagIDs)[0]
					if _, ok := tagsToExcludeMap[tagID]; ok {
						continue
					}
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
				query.excludeTagIDs = tagsToExclude
				actualResult, err := svc.processSummaryQuery(ctx, &query)
				So(err, ShouldBeNil)
				So(len(actualResult), ShouldEqual, len(expectedResults))
				for i, actualSummary := range actualResult {
					expectedSummary := expectedResults[i]
					So(expectedSummary, ShouldResemble, actualSummary)
				}
			})

			Convey("It should subtract refunds from expense", func() {
				expectedByTagID := make(map[int]*summaryDTO)
				expectedResults := []summaryDTO{}

				rndTagIncome := ldtesting.TrxRndTag(md.TagIDs)
				trxref1 := ldtesting.NewTransaction(rndTagIncome, trxDate, rndAcc, ldtesting.TrxRefund)
				trxref2 := ldtesting.NewTransaction(rndTagIncome, trxDate, rndAcc, ldtesting.TrxRefund)
				ldtesting.SetupTransactions(DB, []ldtesting.Transaction{*trxref1, *trxref2})

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

				refTag1 := tags.GetTagIDsFromString(trxref1.TagIDs)[0]
				refTag2 := tags.GetTagIDsFromString(trxref2.TagIDs)[0]
				expectedByTagID[refTag1].Amount -= trxref1.Amount
				expectedByTagID[refTag2].Amount -= trxref2.Amount

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

			Convey("It should not include transactions from other ledgers", func() {
			})
		})

		Convey("When account currency is other than ledger default", func() {
			Convey("It should convert the amount to default currency prior to calculation", func() {

			})
		})
	})
}
