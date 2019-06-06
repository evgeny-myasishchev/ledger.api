package ledgers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"ledger.api/pkg/core/router"

	"github.com/icrowley/fake"
	uuid "github.com/satori/go.uuid"
	. "github.com/smartystreets/goconvey/convey"
	"ledger.api/pkg/internal/ldtesting"
)

type methodCall struct {
	input  interface{}
	result interface{}
}

type mockQueryService struct {
	processUserLedgersQueryCalls []methodCall
}

type ctxKey string

const errorFnKey ctxKey = "error-fn"

func (svc *mockQueryService) processUserLedgersQuery(ctx context.Context, query *userLedgersQuery) ([]ledgerDTO, error) {
	result := []ledgerDTO{
		ledgerDTO{LedgerID: uuid.NewV4().String(), Name: fake.Word(), CurrencyCode: fake.CurrencyCode()},
		ledgerDTO{LedgerID: uuid.NewV4().String(), Name: fake.Word(), CurrencyCode: fake.CurrencyCode()},
		ledgerDTO{LedgerID: uuid.NewV4().String(), Name: fake.Word(), CurrencyCode: fake.CurrencyCode()},
	}
	svc.processUserLedgersQueryCalls = append(svc.processUserLedgersQueryCalls, methodCall{
		input:  []interface{}{query},
		result: result,
	})
	failHandler := ctx.Value(errorFnKey)
	if failHandler != nil {
		return nil, failHandler.(func() error)()
	}
	return result, nil
}

func setupRouter() (*mockQueryService, router.Router) {
	svc := mockQueryService{processUserLedgersQueryCalls: []methodCall{}}
	appRouter := router.CreateRouter()
	SetupRoutes(appRouter, &svc)
	return &svc, appRouter
}

func TestCreateRoute(t *testing.T) {
	svc, router := setupRouter()
	Convey("Given ledger routes", t, func() {
		recorder := httptest.NewRecorder()

		Convey("When route is GET index", func() {
			req := ldtesting.NewRequest("GET", "/v2/ledgers", ldtesting.WithScopeClaim("read:ledgers"))
			router.ServeHTTP(recorder, req)

			Convey("It should respond with user ledgers fetched via query service", func() {
				So(len(svc.processUserLedgersQueryCalls), ShouldEqual, 1)
				queryCall := svc.processUserLedgersQueryCalls[0]
				defaultQuery := &userLedgersQuery{}
				actualQuery := queryCall.input.([]interface{})[0].(*userLedgersQuery)
				So(actualQuery, ShouldResemble, defaultQuery)

				expectedMessage, _ := json.Marshal(queryCall.result)
				So(recorder.Body.String(), ShouldEqual, string(expectedMessage))
				So(recorder.Header().Get("content-type"), ShouldEqual, "application/json")
			})

			Convey("It should respond with 200", func() {
				So(recorder.Code, ShouldEqual, 200)
			})
		})

		// TODO: Fail without scope
	})
}
