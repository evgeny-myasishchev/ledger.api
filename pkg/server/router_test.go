package server

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/icrowley/fake"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRoute(t *testing.T) {

	Convey("Given router", t, func() {
		router := CreateTestRouter()
		recorder := httptest.NewRecorder()

		Convey("When registering routes", func() {
			router.RegisterRoutes(func(r Router) {
				r.GET("/v1/some-resource", func(c Context) (*Response, error) {
					return c.R(JSON{"fake": "string"}), nil
				})
				r.GET("/v1/some-resource/503", func(c Context) (*Response, error) {
					return c.R(JSON{"fake": "string"}).S(503), nil
				})
			})

			Convey("It should invoke provided handler", func() {
				req, _ := http.NewRequest("GET", "/v1/some-resource", nil)
				router.ServeHTTP(recorder, req)

				So(recorder.Code, ShouldEqual, 200)
				expectedMessage, _ := json.Marshal(gin.H{"fake": "string"})
				So(recorder.Body.String(), ShouldEqual, string(expectedMessage))
			})

			Convey("It set custom status code", func() {
				req, _ := http.NewRequest("GET", "/v1/some-resource/503", nil)
				router.ServeHTTP(recorder, req)

				So(recorder.Code, ShouldEqual, 503)
				expectedMessage, _ := json.Marshal(gin.H{"fake": "string"})
				So(recorder.Body.String(), ShouldEqual, string(expectedMessage))
			})
		})

		Convey("When handler returns error", func() {

			Convey("Given standard error", func() {
				router.RegisterRoutes(func(r Router) {
					r.GET("/v1/fail-with-default-error", func(c Context) (*Response, error) {
						return nil, errors.New("Something went very wrong")
					})
				})
				req, _ := http.NewRequest("GET", "/v1/fail-with-default-error", nil)
				router.ServeHTTP(recorder, req)

				Convey("It should respond with default status code", func() {
					So(recorder.Code, ShouldEqual, 500)
				})

				Convey("It should return default error response structure", func() {
					expectedMessage, _ := json.Marshal(JSON{
						"errors": []JSON{
							{
								"status": http.StatusInternalServerError,
								"title":  http.StatusText(http.StatusInternalServerError),
							},
						},
					})
					So(recorder.Body.String(), ShouldEqual, string(expectedMessage))
				})
			})

			Convey("Given http error", func() {
				httpErr := HTTPError{status: rand.Intn(600), title: fake.Sentence()}

				router.RegisterRoutes(func(r Router) {
					r.GET("/v1/fail-with-http-error", func(c Context) (*Response, error) {
						return nil, httpErr
					})
				})

				req, _ := http.NewRequest("GET", "/v1/fail-with-http-error", nil)
				router.ServeHTTP(recorder, req)

				Convey("It should respond with err status code", func() {
					So(recorder.Code, ShouldEqual, httpErr.status)
				})

				Convey("It should return error response structure", func() {
					expectedMessage, _ := json.Marshal(JSON{
						"errors": []JSON{
							{
								"status": httpErr.status,
								"title":  httpErr.title,
							},
						},
					})
					So(recorder.Body.String(), ShouldEqual, string(expectedMessage))
				})
			})
		})
	})
}
