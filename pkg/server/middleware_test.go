package server

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/icrowley/fake"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMiddleware(t *testing.T) {
	Convey("Given middleware", t, func() {
		Convey("When NewRequestIDMiddleware is used", func() {
			middleware := NewRequestIDMiddleware()
			req, _ := http.NewRequest("GET", "/v1/some-resource", nil)
			context := &Context{req: req}

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
