package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRoutes(t *testing.T) {
	router := gin.New()
	RegisterRoutes(router)
	Convey("Given app routes", t, func() {
		recorder := httptest.NewRecorder()
		Convey("When route is healthcheck", func() {
			req, _ := http.NewRequest("GET", "/v1/healthcheck/ping", nil)
			router.ServeHTTP(recorder, req)

			Convey("It should respond with 200", func() {
				So(recorder.Code, ShouldEqual, 200)
			})

			Convey("It should respond with ping", func() {
				expectedMessage, _ := json.Marshal(gin.H{"message": "pong"})
				So(recorder.Body.String(), ShouldEqual, string(expectedMessage))
			})
		})
	})
}