package ledgers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"ledger.api/pkg/server"
)

func TestRoutes(t *testing.T) {
	router := server.
		CreateTestRouter().
		RegisterRoutes(Routes)
	Convey("Given ledger routes", t, func() {
		recorder := httptest.NewRecorder()
		Convey("When route is POST create", func() {
			req, _ := http.NewRequest("POST", "/v2/ledgers", nil)
			router.ServeHTTP(recorder, req)

			Convey("It should respond with 200", func() {
				So(recorder.Code, ShouldEqual, 200)
			})
		})

		Convey("When route is GET index", func() {
			req, _ := http.NewRequest("GET", "/v2/ledgers", nil)
			router.ServeHTTP(recorder, req)

			Convey("It should respond with 200", func() {
				So(recorder.Code, ShouldEqual, 200)
			})
		})
	})
}
