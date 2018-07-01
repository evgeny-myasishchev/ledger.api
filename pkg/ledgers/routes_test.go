package ledgers

import (
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"ledger.api/pkg/internal/ledgertesting"
	"ledger.api/pkg/server"
)

func SetupRouter() *server.HTTPApp {

	return server.
		CreateHTTPApp(server.HTTPAppConfig{Env: "test"}).
		RegisterRoutes(CreateRoutes(service))
}

func TestCreateRoute(t *testing.T) {
	router := SetupRouter()
	Convey("Given ledger routes", t, func() {
		recorder := httptest.NewRecorder()
		Convey("When route is POST create", func() {
			req := ledgertesting.NewRequest("POST", "/v2/ledgers",
				ledgertesting.WithScopeClaim("read:ledgers write:ledgers"))
			router.CreateHandler().ServeHTTP(recorder, req)

			SkipConvey("It should save the ledger", func() {

			})

			Convey("It should respond with 200", func() {
				So(recorder.Code, ShouldEqual, 200)
			})

			SkipConvey("It should respond with ledger details", func() {

			})
		})

		Convey("When route is GET index", func() {
			req := ledgertesting.NewRequest(
				"GET",
				"/v2/ledgers",
				ledgertesting.WithScopeClaim("read:ledgers write:ledgers"))
			router.CreateHandler().ServeHTTP(recorder, req)

			Convey("It should respond with 200", func() {
				So(recorder.Code, ShouldEqual, 200)
			})
		})
	})
}
