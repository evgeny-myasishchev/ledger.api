package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/jsonapi"
	"github.com/icrowley/fake"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBindingAndValidation(t *testing.T) {

	Convey("Given model binding", t, func() {
		app := CreateHTTPApp(HTTPAppConfig{Env: "test"})
		recorder := httptest.NewRecorder()

		type Person struct {
			ID        int    `jsonapi:"primary,persons"`
			FirstName string `jsonapi:"attr,firstName" validate:"required"`
			LastName  string `jsonapi:"attr,lastName" validate:"required"`
		}

		var receivedPerson Person
		app.RegisterRoutes(func(r *Router) {
			r.POST("/v1/persons", func(req *http.Request, h *HandlerToolkit) (*Response, error) {
				err := h.Bind(req, &receivedPerson)
				h.Logger.Infof("Bound person %v %v", err, receivedPerson)
				return h.JSON(nil), err
			})
		})

		handler := app.CreateHandler()

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
			handler.ServeHTTP(recorder, req)

			Convey("It should respond with ok", func() {
				So(recorder.Code, ShouldEqual, 200)
			})

			Convey("It should bind the model data in jsonapi format", func() {
				So(receivedPerson, ShouldResemble, validPerson)
			})
		})

		Convey("When corrupted request body is submitted", func() {
			req, _ := http.NewRequest("POST", "/v1/persons", bytes.NewBufferString("Some crap"))
			handler.ServeHTTP(recorder, req)

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
			handler.ServeHTTP(recorder, req)

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
