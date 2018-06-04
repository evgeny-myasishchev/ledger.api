package server

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

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
				r.GET("/v1/some-resource", func(req *http.Request, h *HandlerToolkit) (*Response, error) {
					return h.JSON(JSON{"fake": "string"}), nil
				})
				r.GET("/v1/some-resource/503", func(req *http.Request, h *HandlerToolkit) (*Response, error) {
					return h.JSON(JSON{"fake": "string"}).Status(503), nil
				})
			})
			handler := router.CreateHandler()

			Convey("It should invoke provided handler", func() {
				req, _ := http.NewRequest("GET", "/v1/some-resource", nil)
				handler.ServeHTTP(recorder, req)

				So(recorder.Code, ShouldEqual, 200)
				expectedMessage, _ := json.Marshal(JSON{"fake": "string"})
				So(recorder.Body.String(), ShouldEqual, string(expectedMessage))
			})

			Convey("It set custom status code", func() {
				req, _ := http.NewRequest("GET", "/v1/some-resource/503", nil)
				handler.ServeHTTP(recorder, req)

				So(recorder.Code, ShouldEqual, 503)
				expectedMessage, _ := json.Marshal(JSON{"fake": "string"})
				So(recorder.Body.String(), ShouldEqual, string(expectedMessage))
			})
		})

		Convey("When request to unknown route", func() {
			req, _ := http.NewRequest("GET", "/v1/some-resource", nil)
			handler := router.CreateHandler()
			handler.ServeHTTP(recorder, req)

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
					r.GET("/v1/fail-with-default-error", func(req *http.Request, h *HandlerToolkit) (*Response, error) {
						return nil, errors.New("Something went very wrong")
					})
				})
				req, _ := http.NewRequest("GET", "/v1/fail-with-default-error", nil)
				handler := router.CreateHandler()
				handler.ServeHTTP(recorder, req)

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
					r.GET("/v1/fail-with-http-error", func(req *http.Request, h *HandlerToolkit) (*Response, error) {
						return nil, httpErr
					})
				})

				req, _ := http.NewRequest("GET", "/v1/fail-with-http-error", nil)
				handler := router.CreateHandler()
				handler.ServeHTTP(recorder, req)

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
