package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRoute(t *testing.T) {

	Convey("Given router", t, func() {
		router := CreateNewRouter()
		recorder := httptest.NewRecorder()

		Convey("When registering routes", func() {

			router.RegisterRoutes(func(r Router) {
				r.GET("/v1/some-resource", func(c Context) {
					c.JSON(200, JSON{"fake": "string"})
				})
			})

			req, _ := http.NewRequest("GET", "/v1/some-resource", nil)
			router.ServeHTTP(recorder, req)

			Convey("It should invoke provided handler", func() {
				So(recorder.Code, ShouldEqual, 200)
				expectedMessage, _ := json.Marshal(gin.H{"fake": "string"})
				So(recorder.Body.String(), ShouldEqual, string(expectedMessage))
			})
		})
	})
}
