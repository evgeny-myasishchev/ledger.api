package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/icrowley/fake"
	. "github.com/smartystreets/goconvey/convey"
	"ledger.api/pkg/logging"
)

func TestRouteMiddleware(t *testing.T) {

	Convey("Given router middleware", t, func() {
		app := CreateHTTPApp(HTTPAppConfig{Env: "test"})
		recorder := httptest.NewRecorder()

		handlerCalled := false
		app.RegisterRoutes(func(r *Router) {
			r.handle("GET", "/v1/some-resource", func(c *Context) (*Response, error) {
				handlerCalled = true
				return c.R(JSON{"fake": "string"}), nil
			})
			r.handle("GET", "/v1/should-abort", func(c *Context) (*Response, error) {
				return c.R(JSON{"fake": "string"}), nil
			})
		})

		Convey("When middleware is registered", func() {
			middlewareInvoked := false
			app.Use(func(next http.HandlerFunc) http.HandlerFunc {
				return func(w http.ResponseWriter, req *http.Request) {
					logger := logging.FromContext(req.Context())
					logger.Infof("Processing req %v with test middleware", req.URL.Path)
					middlewareInvoked = true
					if req.URL.Path == "/v1/should-abort" {
						w.WriteHeader(500)
						buffer, err := json.Marshal(JSON{"aborted": "true"})
						if err != nil {
							panic(err)
						}
						if _, err := w.Write(buffer); err != nil {
							panic(err)
						}
						return
					}
					next(w, req)
				}
			})

			Convey("It should invoke middleware prior to route handler", func() {
				req, _ := http.NewRequest("GET", "/v1/some-resource", nil)
				handler := app.CreateHandler()
				handler.ServeHTTP(recorder, req)
				So(middlewareInvoked, ShouldBeTrue)
			})

			Convey("It should invoke multiple middlewares in order of registering", func() {
				callCount := 0
				app.RegisterRoutes(func(r *Router) {
					r.handle("GET", "/v1/some-ordered-route", func(c *Context) (*Response, error) {
						c.Logger.Info("Processing actual route handler")
						So(callCount, ShouldEqual, 3)
						callCount++
						return c.R(JSON{"fake": "string"}), nil
					})
				})

				app.Use(func(next http.HandlerFunc) http.HandlerFunc {
					return func(w http.ResponseWriter, req *http.Request) {
						logger := logging.FromContext(req.Context())
						logger.Info("Processing mw 0")
						So(callCount, ShouldEqual, 0)
						callCount++
						next(w, req)
					}
				})
				app.Use(func(next http.HandlerFunc) http.HandlerFunc {
					return func(w http.ResponseWriter, req *http.Request) {
						logger := logging.FromContext(req.Context())
						logger.Info("Processing mw 1")
						So(callCount, ShouldEqual, 1)
						callCount++
						next(w, req)
					}
				})
				app.Use(func(next http.HandlerFunc) http.HandlerFunc {
					return func(w http.ResponseWriter, req *http.Request) {
						logger := logging.FromContext(req.Context())
						logger.Info("Processing mw 2")
						So(callCount, ShouldEqual, 2)
						callCount++
						next(w, req)
					}
				})

				handler := app.CreateHandler()
				req, _ := http.NewRequest("GET", "/v1/some-ordered-route", nil)
				handler.ServeHTTP(recorder, req)
				So(middlewareInvoked, ShouldBeTrue)
				So(callCount, ShouldEqual, 4)
			})

			Convey("It not invoke route handler if middleware aborts the invocation", func() {
				req, _ := http.NewRequest("GET", "/v1/should-abort", nil)
				handler := app.CreateHandler()
				handler.ServeHTTP(recorder, req)
				So(middlewareInvoked, ShouldBeTrue)
				So(handlerCalled, ShouldBeFalse)
				So(recorder.Code, ShouldEqual, 500)
			})
		})
	})
}

func TestNewRequestIDMiddleware(t *testing.T) {
	Convey("Given RequestIDMiddleware", t, func() {
		recorder := httptest.NewRecorder()
		Convey("When NewRequestIDMiddleware is used", func() {
			middleware := NewRequestIDMiddleware
			context := logging.CreateContext(context.Background(), logging.NewTestLogger())
			req, _ := http.NewRequest("GET", "/v1/some-resource", nil)
			req = req.WithContext(context)

			Convey("It should generate a new request id", func() {
				var requestID string
				middleware(func(w http.ResponseWriter, mwReq *http.Request) {
					requestID = RequestIDVAlue(mwReq.Context())
				})(recorder, req)
				So(requestID, ShouldNotBeEmpty)
			})

			Convey("It should use a request id from X-Request-ID header", func() {
				reqID := fmt.Sprintf("req-id-%v", fake.Characters())
				req.Header.Add("X-Request-ID", reqID)
				var actualRequestID string
				middleware(func(w http.ResponseWriter, mwReq *http.Request) {
					actualRequestID = RequestIDVAlue(mwReq.Context())
				})(recorder, req)
				So(actualRequestID, ShouldEqual, reqID)
			})

			Convey("It should call next", func() {
				nextCalled := false
				middleware(func(w http.ResponseWriter, mwReq *http.Request) {
					nextCalled = true
				})(recorder, req)
				So(nextCalled, ShouldBeTrue)
			})
		})
	})
}
