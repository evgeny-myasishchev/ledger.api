package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	tst "ledger.api/pkg/internal/testing"

	"github.com/auth0-community/go-auth0"

	"github.com/bxcodec/faker/v3"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"
	"gopkg.in/square/go-jose.v2/jwt"

	"ledger.api/pkg/core/router"
)

type mockValidator struct {
	mock.Mock
	initClaims func(...interface{})
}

func (v *mockValidator) ValidateRequest(r *http.Request) (*jwt.JSONWebToken, error) {
	args := v.Called(r)
	return args.Get(0).(*jwt.JSONWebToken), args.Error(1)
}
func (v *mockValidator) Claims(r *http.Request, token *jwt.JSONWebToken, values ...interface{}) error {
	args := v.Called(r, token)
	if v.initClaims != nil {
		v.initClaims(values...)
	}
	return args.Error(0)
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
					v.On("ValidateRequest", req).Return((*jwt.JSONWebToken)(nil), auth0.ErrTokenNotFound)
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
		func() testCase {
			v := &mockValidator{}
			return testCase{
				name: "valid token",
				args: args{setup: []MiddlewareOpt{withValidator(v)}},
				run: func(t *testing.T, mw router.MiddlewareFunc) {

					req := httptest.NewRequest("GET", "/some-path", nil)
					recorder := httptest.NewRecorder()
					nextCalled := false
					expectedScope := "scope-" + faker.Word()

					token := &jwt.JSONWebToken{}
					v.On("ValidateRequest", req).Return(token, nil)
					v.On("Claims", req, token, mock.Anything).Return(nil)
					v.initClaims = func(targets ...interface{}) {
						if !assert.Len(t, targets, 1) {
							return
						}
						claims := targets[0].(*LedgerClaims)
						claims.Scope = expectedScope
					}

					var actualClaims *LedgerClaims
					next := func(w http.ResponseWriter, req *http.Request) {
						actualClaims = ClaimsFromContext(req.Context())
						nextCalled = true
					}
					mw(next)(recorder, req)
					assert.True(t, nextCalled)
					if !assert.NotNil(t, actualClaims) {
						return
					}
					assert.Equal(t, expectedScope, actualClaims.Scope)
					v.AssertExpectations(t)
				},
			}
		},
		func() testCase {
			v := &mockValidator{}
			return testCase{
				name: "invalid token",
				args: args{setup: []MiddlewareOpt{withValidator(v)}},
				run: func(t *testing.T, mw router.MiddlewareFunc) {

					req := httptest.NewRequest("GET", "/some-path", nil)
					recorder := httptest.NewRecorder()
					nextCalled := false

					err := errors.New(faker.Sentence())
					v.On("ValidateRequest", req).Return((*jwt.JSONWebToken)(nil), err)

					next := func(w http.ResponseWriter, req *http.Request) {
						nextCalled = true
					}
					mw(next)(recorder, req)
					if !assert.False(t, nextCalled) {
						return
					}
					v.AssertExpectations(t)

					assert.Equal(t, http.StatusUnauthorized, recorder.Code)

					var httpError router.HTTPError
					if !tst.JSONUnmarshalReader(t, recorder.Body, &httpError) {
						return
					}
					assert.Equal(t, router.HTTPError{
						StatusCode: http.StatusUnauthorized,
						Status:     http.StatusText(http.StatusUnauthorized),
						Message:    "Token validation failed",
					}, httpError)
				},
			}
		},
		func() testCase {
			v := &mockValidator{}
			return testCase{
				name: "invalid token claims",
				args: args{setup: []MiddlewareOpt{withValidator(v)}},
				run: func(t *testing.T, mw router.MiddlewareFunc) {

					req := httptest.NewRequest("GET", "/some-path", nil)
					recorder := httptest.NewRecorder()
					nextCalled := false

					err := errors.New(faker.Sentence())
					token := &jwt.JSONWebToken{}
					v.On("ValidateRequest", req).Return(token, nil)
					v.On("Claims", req, token, mock.Anything).Return(err)

					next := func(w http.ResponseWriter, req *http.Request) {
						nextCalled = true
					}
					mw(next)(recorder, req)
					if !assert.False(t, nextCalled) {
						return
					}
					v.AssertExpectations(t)

					assert.Equal(t, http.StatusUnauthorized, recorder.Code)

					var httpError router.HTTPError
					if !tst.JSONUnmarshalReader(t, recorder.Body, &httpError) {
						return
					}
					assert.Equal(t, router.HTTPError{
						StatusCode: http.StatusUnauthorized,
						Status:     http.StatusText(http.StatusUnauthorized),
						Message:    "Bad token",
					}, httpError)
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
