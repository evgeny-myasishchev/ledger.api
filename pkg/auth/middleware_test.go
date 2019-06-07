package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"
	"gopkg.in/square/go-jose.v2/jwt"

	"ledger.api/pkg/core/router"
)

type mockValidator struct {
	mock.Mock
}

func (v *mockValidator) ValidateRequest(r *http.Request) (*jwt.JSONWebToken, error) {
	return nil, nil
}
func (v *mockValidator) Claims(r *http.Request, token *jwt.JSONWebToken, values ...interface{}) error {
	return nil
}

func TestMiddleware(t *testing.T) {
	type args struct {
		setup []MiddlewareOpt
	}
	type testCase struct {
		name string
		args args
		run  func(t *testing.T, mw router.MiddlewareFunc)
	}
	tests := []func() testCase{
		func() testCase {
			v := &mockValidator{}
			return testCase{
				name: "no token",
				args: args{setup: []MiddlewareOpt{withValidator(v)}},
				run: func(t *testing.T, mw router.MiddlewareFunc) {
					req := httptest.NewRequest("GET", "/some-path", nil)
					recorder := httptest.NewRecorder()
					nextCalled := false
					next := func(w http.ResponseWriter, req *http.Request) {
						nextCalled = true
					}
					mw(next)(recorder, req)
					assert.True(t, nextCalled)
				},
			}
		},
	}
	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			mw := NewMiddleware(tt.args.setup...)
			tt.run(t, mw)
		})
	}
}
