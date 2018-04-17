package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRouteMiddleware(t *testing.T) {

	Convey("Given router", t, func() {
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
			app.Use(func(ctx *Context, next HandlerFunc) (*Response, error) {
				ctx.Logger.Infof("Processing req %v with test middleware", ctx.req.URL.Path)
				middlewareInvoked = true
				if ctx.req.URL.Path == "/v1/should-abort" {
					return ctx.R(JSON{"aborted": "true"}).S(500), nil
				}
				return next(ctx)
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

				app.Use(func(ctx *Context, next HandlerFunc) (*Response, error) {
					ctx.Logger.Info("Processing mw 0")
					So(callCount, ShouldEqual, 0)
					callCount++
					return next(ctx)
				})
				app.Use(func(ctx *Context, next HandlerFunc) (*Response, error) {
					ctx.Logger.Info("Processing mw 1")
					So(callCount, ShouldEqual, 1)
					callCount++
					return next(ctx)
				})

				req, _ := http.NewRequest("GET", "/v1/some-ordered-route", nil)
				app.ServeHTTP(recorder, req)
				So(middlewareInvoked, ShouldBeTrue)
				So(callCount, ShouldEqual, 3)
			})

			Convey("It should not invoke middleware for unknown routes", func() {
				req, _ := http.NewRequest("GET", "/v1/unknown-route", nil)
				app.ServeHTTP(recorder, req)
				So(middlewareInvoked, ShouldBeFalse)
				So(recorder.Code, ShouldEqual, 404)
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
