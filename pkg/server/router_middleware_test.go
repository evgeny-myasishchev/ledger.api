package server

import (
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
		logger := logging.NewTestLogger()

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
			app.Use2(func(w http.ResponseWriter, req *http.Request, next func()) {
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
				next()
			})

			Convey("It should invoke middleware prior to route handler", func() {
				req, _ := http.NewRequest("GET", "/v1/some-resource", nil)
				app.ServeHTTP(recorder, req)
				So(middlewareInvoked, ShouldBeTrue)
			})

			Convey("It should invoke multiple middlewares in order of registering", func() {
				callCount := 0
				app.RegisterRoutes(func(r *Router) {
					r.handle("GET", "/v1/some-ordered-route", func(c *Context) (*Response, error) {
						c.Logger.Info("Processing actual route handler")
						So(callCount, ShouldEqual, 2)
						callCount++
						return c.R(JSON{"fake": "string"}), nil
					})
				})

				app.Use2(func(w http.ResponseWriter, req *http.Request, next func()) {
					logger.Info("Processing mw 0")
					So(callCount, ShouldEqual, 0)
					callCount++
					next()
				})
				app.Use2(func(w http.ResponseWriter, req *http.Request, next func()) {
					logger.Info("Processing mw 1")
					So(callCount, ShouldEqual, 1)
					callCount++
					next()
				})

				req, _ := http.NewRequest("GET", "/v1/some-ordered-route", nil)
				app.ServeHTTP(recorder, req)
				So(middlewareInvoked, ShouldBeTrue)
				So(callCount, ShouldEqual, 3)
			})

			Convey("It not invoke route handler if middleware aborts the invocation", func() {
				req, _ := http.NewRequest("GET", "/v1/should-abort", nil)
				app.ServeHTTP(recorder, req)
				So(middlewareInvoked, ShouldBeTrue)
				So(handlerCalled, ShouldBeFalse)
				So(recorder.Code, ShouldEqual, 500)
			})
		})
	})
}

func TestNewRequestIDMiddleware(t *testing.T) {
	Convey("Given RequestIDMiddleware", t, func() {
		Convey("When NewRequestIDMiddleware is used", func() {
			middleware := NewRequestIDMiddleware()
			req, _ := http.NewRequest("GET", "/v1/some-resource", nil)
			context := &Context{req: req, Logger: logging.NewTestLogger()}

			Convey("It should generate a new request id", func() {
				middleware(context, func(ctx *Context) (*Response, error) {
					return nil, nil
				})
				So(context.requestID, ShouldNotBeEmpty)
			})

			Convey("It should use a request id from X-Request-ID header", func() {
				reqID := fmt.Sprintf("req-id-%v", fake.Characters())
				req.Header.Add("X-Request-ID", reqID)
				middleware(context, func(ctx *Context) (*Response, error) {
					return nil, nil
				})
				So(context.requestID, ShouldEqual, reqID)
			})

			Convey("It should call next", func() {
				nextCalled := false
				middleware(context, func(ctx *Context) (*Response, error) {
					nextCalled = true
					return nil, nil
				})
				So(nextCalled, ShouldBeTrue)
			})
		})
	})
}
