package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/jsonapi"
	"github.com/icrowley/fake"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRoute(t *testing.T) {

	Convey("Given router", t, func() {
		router := CreateHTTPApp(HTTPAppConfig{Env: "test"})
		recorder := httptest.NewRecorder()

		Convey("When registering routes", func() {
			router.RegisterRoutes(func(r *Router) {
				r.GET("/v1/some-resource", func(c *Context) (*Response, error) {
					return c.R(JSON{"fake": "string"}), nil
				})
				r.GET("/v1/some-resource/503", func(c *Context) (*Response, error) {
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

		Convey("When request to unknown route", func() {
			req, _ := http.NewRequest("GET", "/v1/some-resource", nil)
			router.ServeHTTP(recorder, req)

			Convey("It should respond 404 status", func() {
				So(recorder.Code, ShouldEqual, 404)
			})
			Convey("It should respond with consistent error body", func() {
				expectedMessage := map[string]interface{}{
					"errors": []interface{}{
						map[string]interface{}{
							"status": strconv.Itoa(http.StatusNotFound),
							"title":  http.StatusText(http.StatusNotFound),
						},
					},
				}
				var actualMessage map[string]interface{}
				if err := json.Unmarshal(recorder.Body.Bytes(), &actualMessage); err != nil {
					panic(err)
				}
				So(actualMessage, ShouldResemble, expectedMessage)
			})
		})

		Convey("When handler returns error", func() {

			Convey("Given standard error", func() {
				router.RegisterRoutes(func(r *Router) {
					r.GET("/v1/fail-with-default-error", func(c *Context) (*Response, error) {
						return nil, errors.New("Something went very wrong")
					})
				})
				req, _ := http.NewRequest("GET", "/v1/fail-with-default-error", nil)
				router.ServeHTTP(recorder, req)

				Convey("It should respond with default status code", func() {
					So(recorder.Code, ShouldEqual, 500)
				})

				Convey("It should return default error response structure", func() {
					expectedMessage := map[string]interface{}{
						"errors": []interface{}{
							map[string]interface{}{
								"status": strconv.Itoa(http.StatusInternalServerError),
								"title":  http.StatusText(http.StatusInternalServerError),
							},
						},
					}
					var actualMessage map[string]interface{}
					if err := json.Unmarshal(recorder.Body.Bytes(), &actualMessage); err != nil {
						panic(err)
					}
					So(actualMessage, ShouldResemble, expectedMessage)
				})
			})

			Convey("Given http error", func() {
				httpErr := HTTPError{
					Status: rand.Intn(600),
					Errors: []*jsonapi.ErrorObject{
						{
							Status: strconv.Itoa(rand.Intn(500)),
							Title:  fake.Sentence(),
						},
					},
				}

				router.RegisterRoutes(func(r *Router) {
					r.GET("/v1/fail-with-http-error", func(c *Context) (*Response, error) {
						return nil, httpErr
					})
				})

				req, _ := http.NewRequest("GET", "/v1/fail-with-http-error", nil)
				router.ServeHTTP(recorder, req)

				Convey("It should respond with err status code", func() {
					So(recorder.Code, ShouldEqual, httpErr.Status)
				})

				Convey("It should return error response structure", func() {
					expectedMessage := map[string]interface{}{
						"errors": []interface{}{
							map[string]interface{}{
								"status": httpErr.Errors[0].Status,
								"title":  httpErr.Errors[0].Title,
							},
						},
					}
					var actualMessage map[string]interface{}
					if err := json.Unmarshal(recorder.Body.Bytes(), &actualMessage); err != nil {
						panic(err)
					}
					So(actualMessage, ShouldResemble, expectedMessage)
				})
			})
		})
	})
}

func TestBindingAndValidation(t *testing.T) {

	Convey("Given model binding", t, func() {
		router := CreateHTTPApp(HTTPAppConfig{Env: "test"})
		recorder := httptest.NewRecorder()

		type Person struct {
			ID        int    `jsonapi:"primary,persons"`
			FirstName string `jsonapi:"attr,firstName" validate:"required"`
			LastName  string `jsonapi:"attr,lastName" validate:"required"`
		}

		var receivedPerson Person
		router.RegisterRoutes(func(r *Router) {
			r.POST("/v1/persons", func(c *Context) (*Response, error) {
				err := c.Bind(&receivedPerson)
				c.Logger.Infof("Bound person %v %v", err, receivedPerson)
				return c.R(nil), err
			})
		})

		Convey("When valid request body is submitted", func() {
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

		Convey("When corrupted request body is submitted", func() {
			req, _ := http.NewRequest("POST", "/v1/persons", bytes.NewBufferString("Some crap"))
			router.ServeHTTP(recorder, req)

			Convey("It should fail with 500 error", func() {
				So(recorder.Code, ShouldEqual, 500)
			})

			Convey("It should include standard 500 error details", func() {
				expectedMessage := map[string]interface{}{
					"errors": []interface{}{
						map[string]interface{}{
							"status": "500",
							"title":  "Internal Server Error",
						},
					},
				}
				var actualMessage map[string]interface{}
				if err := json.Unmarshal(recorder.Body.Bytes(), &actualMessage); err != nil {
					panic(err)
				}

				So(actualMessage, ShouldResemble, expectedMessage)
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
				expectedMessage := map[string]interface{}{
					"errors": []interface{}{
						map[string]interface{}{
							"status": "400",
							"title":  "Validation error",
							"detail": "Field 'Person.FirstName' validation failed on 'required' tag",
						},
						map[string]interface{}{
							"status": "400",
							"title":  "Validation error",
							"detail": "Field 'Person.LastName' validation failed on 'required' tag",
						},
					},
				}
				var actualMessage map[string]interface{}
				if err := json.Unmarshal(recorder.Body.Bytes(), &actualMessage); err != nil {
					panic(err)
				}
				So(actualMessage, ShouldResemble, expectedMessage)
			})
		})
	})
}
