package transactions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/satori/go.uuid"

	"github.com/icrowley/fake"

	. "github.com/smartystreets/goconvey/convey"
	"ledger.api/pkg/internal/ldtesting"
	"ledger.api/pkg/server"
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

type methodCall struct {
	input  interface{}
	result interface{}
}

type mockQueryService struct {
	processSummaryQueryCalls []methodCall
}

type ctxKey string

const errorFnKey ctxKey = "error-fn"

func (svc *mockQueryService) processSummaryQuery(ctx context.Context, query *summaryQuery) ([]summaryDTO, error) {
	result := []summaryDTO{
		summaryDTO{TagID: rnd.Int(), TagName: fake.Word(), Amount: rnd.Int()},
		summaryDTO{TagID: rnd.Int(), TagName: fake.Word(), Amount: rnd.Int()},
		summaryDTO{TagID: rnd.Int(), TagName: fake.Word(), Amount: rnd.Int()},
	}
	svc.processSummaryQueryCalls = append(svc.processSummaryQueryCalls, methodCall{
		input:  []interface{}{query},
		result: result,
	})
	failHandler := ctx.Value(errorFnKey)
	if failHandler != nil {
		return nil, failHandler.(func() error)()
	}
	return result, nil
}

func setupRouter() (*mockQueryService, *server.HTTPApp) {
	svc := mockQueryService{processSummaryQueryCalls: []methodCall{}}
	return &svc, server.
		CreateHTTPApp(server.HTTPAppConfig{Env: "test"}).
		RegisterRoutes(CreateRoutes(&svc))
}

func TestTransactionsRoutes(t *testing.T) {
	Convey("Given transactions routes", t, func() {
		svc, router := setupRouter()
		recorder := httptest.NewRecorder()
		ledgerID := uuid.NewV4().String()
		typ := fake.Word()
		Convey("When route is processSummaryQuery", func() {
			path := fmt.Sprintf("/v2/ledgers/%v/transactions/%v/summary", ledgerID, typ)

			Convey("And user is authorized", func() {
				req := ldtesting.NewRequest("GET", path, ldtesting.WithScopeClaim("read:transactions"))
				Convey("It should process query and return summary data", func() {
					router.CreateHandler().ServeHTTP(recorder, req)
					So(recorder.Code, ShouldEqual, 200)
					So(len(svc.processSummaryQueryCalls), ShouldEqual, 1)
					queryCall := svc.processSummaryQueryCalls[0]
					defaultQuery := newSummaryQuery(ledgerID, typ)

					actualQuery := queryCall.input.([]interface{})[0].(*summaryQuery)
					So(actualQuery.from.Unix(), ShouldAlmostEqual, defaultQuery.from.Unix())
					So(actualQuery.to.Unix(), ShouldAlmostEqual, defaultQuery.to.Unix())
					actualQuery.from = defaultQuery.from
					actualQuery.to = defaultQuery.to
					So(actualQuery, ShouldResemble, defaultQuery)

					expectedMessage, _ := json.Marshal(queryCall.result)
					So(recorder.Body.String(), ShouldEqual, string(expectedMessage))
					So(recorder.Header().Get("content-type"), ShouldEqual, "application/json")
				})

				Convey("It should use query string params", func() {
					from := ldtesting.RandomDate()
					to := ldtesting.RandomDate()
					qs := url.Values{}
					qs.Add("from", from.Format(time.RFC3339))
					qs.Add("to", to.Format(time.RFC3339))

					excludeTagIDs := []string{fake.Word(), fake.Word(), fake.Word()}
					qs.Add("excludeTagIDs", strings.Join(excludeTagIDs, ","))

					url := fmt.Sprintf("/v2/ledgers/%v/transactions/%v/summary?%v", ledgerID, typ, qs.Encode())
					req := ldtesting.NewRequest("GET", url, ldtesting.WithScopeClaim("read:transactions"))
					router.CreateHandler().ServeHTTP(recorder, req)
					So(recorder.Code, ShouldEqual, 200)
					So(len(svc.processSummaryQueryCalls), ShouldEqual, 1)
					queryCall := svc.processSummaryQueryCalls[0]

					inputQuery := queryCall.input.([]interface{})[0].(*summaryQuery)
					So(inputQuery.from, ShouldNotBeNil)
					So(inputQuery.from.Format(time.RFC3339), ShouldEqual, from.Format(time.RFC3339))

					So(inputQuery.to, ShouldNotBeNil)
					So(inputQuery.to.Format(time.RFC3339), ShouldEqual, to.Format(time.RFC3339))
					So(inputQuery.excludeTagIDs, ShouldResemble, excludeTagIDs)
				})

				Convey("It should respond with error if query fails", func() {
					failCtx := context.WithValue(req.Context(), errorFnKey, func() error {
						return errors.New(fake.Sentence())
					})
					router.CreateHandler().ServeHTTP(recorder, req.WithContext(failCtx))
					So(recorder.Code, ShouldEqual, 500)
					So(len(svc.processSummaryQueryCalls), ShouldEqual, 1)
				})
			})

			Convey("And user is not authorized", func() {
				req := ldtesting.NewRequest("GET", path, ldtesting.WithScopeClaim("none"))
				Convey("It should reject with 403", func() {
					router.CreateHandler().ServeHTTP(recorder, req)
					So(recorder.Code, ShouldEqual, 403)
				})
			})
		})
	})
}
