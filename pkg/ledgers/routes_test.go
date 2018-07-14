package ledgers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/satori/go.uuid"

	"github.com/icrowley/fake"
	. "github.com/smartystreets/goconvey/convey"
	"ledger.api/pkg/internal/ldtesting"
	"ledger.api/pkg/server"
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

func setupRouter() (*mockQueryService, *server.HTTPApp) {
	svc := mockQueryService{processUserLedgersQueryCalls: []methodCall{}}
	return &svc, server.
		CreateHTTPApp(server.HTTPAppConfig{Env: "test"}).
		RegisterRoutes(CreateRoutes(&svc))
}

func TestCreateRoute(t *testing.T) {
	svc, router := setupRouter()
	Convey("Given ledger routes", t, func() {
		recorder := httptest.NewRecorder()

		Convey("When route is GET index", func() {
			req := ldtesting.NewRequest(
				"GET",
				"/v2/ledgers",
				ldtesting.WithScopeClaim("read:ledgers"))
			router.CreateHandler().ServeHTTP(recorder, req)

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
	})
}
