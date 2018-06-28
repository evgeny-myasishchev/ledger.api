package transactions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/satori/go.uuid"

	"github.com/icrowley/fake"

	. "github.com/smartystreets/goconvey/convey"
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
			req, _ := http.NewRequest("GET", fmt.Sprintf("/v2/ledgers/%v/transactions/%v/summary", ledgerID, typ), nil)

			Convey("And user is authenticated", func() {
				Convey("It should process query and return summary data", func() {
					router.CreateHandler().ServeHTTP(recorder, req)
					So(recorder.Code, ShouldEqual, 200)
					So(len(svc.processSummaryQueryCalls), ShouldEqual, 1)
					queryCall := svc.processSummaryQueryCalls[0]
					So(queryCall.input, ShouldResemble, []interface{}{&summaryQuery{
						ledgerID: ledgerID,
						typ:      typ,
					}})

					expectedMessage, _ := json.Marshal(queryCall.result)
					So(recorder.Body.String(), ShouldEqual, string(expectedMessage))
					So(recorder.Header().Get("content-type"), ShouldEqual, "application/json")
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
		})
	})
}
