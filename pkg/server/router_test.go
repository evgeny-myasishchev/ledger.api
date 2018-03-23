package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/jsonapi"
	"github.com/icrowley/fake"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRoute(t *testing.T) {

	Convey("Given router", t, func() {
		router := CreateTestRouter()
		recorder := httptest.NewRecorder()

		SkipConvey("When registering routes", func() {
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

		SkipConvey("When handler returns error", func() {

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

		type Person struct {
			ID        int    `jsonapi:"primary,persons"`
			FirstName string `jsonapi:"attr,firstName" validate:"required"`
			LastName  string `jsonapi:"attr,lastName" validate:"required"`
		}

		var receivedPerson Person
		router.RegisterRoutes(func(r Router) {
			r.POST("/v1/persons", func(c Context) (*Response, error) {
				err := c.Bind(&receivedPerson)
				c.Logger().Infof("Bound person %v %v", err, receivedPerson)
				return c.R(nil), err
			})
		})

		SkipConvey("When valid request body is submitted", func() {
			validPerson := Person{
				ID:        10,
				FirstName: fake.FirstName(),
				LastName:  fake.LastName(),
			}

			data := bytes.NewBuffer(nil)
			if err := jsonapi.MarshalPayload(data, &validPerson); err != nil {
				panic(err)
			}

			req, _ := http.NewRequest("POST", "/v1/persons", data)
			router.ServeHTTP(recorder, req)

			Convey("It should respond with ok", func() {
				So(recorder.Code, ShouldEqual, 200)
			})

			Convey("It should bind the model data in jsonapi format", func() {
				So(receivedPerson, ShouldResemble, validPerson)
			})
		})

		SkipConvey("When corrupted request body is submitted", func() {
			req, _ := http.NewRequest("POST", "/v1/persons", bytes.NewBufferString("Some crap"))
			router.ServeHTTP(recorder, req)

			Convey("It should fail with 500 error", func() {
				So(recorder.Code, ShouldEqual, 500)
			})

			Convey("It should include standard 500 error details", func() {
				expectedMessage, _ := json.Marshal(JSON{
					"errors": []JSON{
						{
							"status": 500,
							"title":  "Internal Server Error",
						},
					},
				})
				So(recorder.Body.String(), ShouldEqual, string(expectedMessage))
			})
		})

		Convey("When invalid request body is submitted", func() {
			invalidPerson := Person{
				ID: 10,
			}

			data := bytes.NewBuffer(nil)
			if err := jsonapi.MarshalPayload(data, &invalidPerson); err != nil {
				panic(err)
			}

			req, _ := http.NewRequest("POST", "/v1/persons", data)
			router.ServeHTTP(recorder, req)

			Convey("It should respond with bad request", func() {
				So(recorder.Code, ShouldEqual, 400)
			})

			Convey("It should respond with error details", func() {
				expectedMessage, _ := json.Marshal(JSON{
					"errors": []JSON{
						{
							"status": 400,
							"title":  "FirstName is required",
						},
						{
							"status": 400,
							"title":  "LastName is required",
						},
					},
				})
				So(recorder.Body.String(), ShouldEqual, string(expectedMessage))
			})
		})
	})
}
