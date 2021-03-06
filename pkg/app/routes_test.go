package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"ledger.api/pkg/server"
)

func TestRoutes(t *testing.T) {
	router := server.
		CreateHTTPApp(server.HTTPAppConfig{Env: "test"}).
		RegisterRoutes(Routes)
	Convey("Given app routes", t, func() {
		recorder := httptest.NewRecorder()
		Convey("When route is healthcheck", func() {
			req, _ := http.NewRequest("GET", "/v2/healthcheck/ping", nil)
			router.CreateHandler().ServeHTTP(recorder, req)

			Convey("It should respond with 200", func() {
				So(recorder.Code, ShouldEqual, 200)
			})

			Convey("It should respond with ping", func() {
				expectedMessage, _ := json.Marshal(server.JSON{"message": "pong"})
				So(recorder.Body.String(), ShouldEqual, string(expectedMessage))
			})
		})
	})
}
