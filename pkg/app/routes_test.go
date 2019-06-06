package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ledger.api/pkg/core/router"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRoutes(t *testing.T) {
	appRouter := router.CreateRouter()
	SetupRoutes(appRouter)
	Convey("Given app routes", t, func() {
		recorder := httptest.NewRecorder()
		Convey("When route is healthcheck", func() {
			req, _ := http.NewRequest("GET", "/v2/healthcheck/ping", nil)
			appRouter.ServeHTTP(recorder, req)

			Convey("It should respond with 200", func() {
				So(recorder.Code, ShouldEqual, 200)
			})

			Convey("It should respond with ping", func() {
				expectedMessage, _ := json.Marshal(map[string]interface{}{"ping": "PONG"})
				So(recorder.Body.String(), ShouldEqual, string(expectedMessage))
			})
		})
	})
}
