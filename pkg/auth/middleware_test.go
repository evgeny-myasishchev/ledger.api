package auth

import (
	"errors"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/auth0-community/go-auth0"
	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/square/go-jose.v2/jwt"
	"ledger.api/pkg/core/router"
	tst "ledger.api/pkg/internal/testing"
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
				args: args{setup: []MiddlewareOpt{WithValidator(v)}},
				run: func(t *testing.T, mw router.MiddlewareFunc) {
					req := httptest.NewRequest("GET", "/some-path", nil)
					v.On("ValidateRequest", req).Return((*jwt.JSONWebToken)(nil), auth0.ErrTokenNotFound)
					recorder := httptest.NewRecorder()
					nextCalled := false
					next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						nextCalled = true
					})
					mw(next).ServeHTTP(recorder, req)
					assert.True(t, nextCalled)
				},
			}
		},
		func() testCase {
			v := &mockValidator{}
			return testCase{
				name: "valid token",
				args: args{setup: []MiddlewareOpt{WithValidator(v)}},
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
					next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						actualClaims = ClaimsFromContext(req.Context())
						nextCalled = true
					})
					mw(next).ServeHTTP(recorder, req)
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
				args: args{setup: []MiddlewareOpt{WithValidator(v)}},
				run: func(t *testing.T, mw router.MiddlewareFunc) {

					req := httptest.NewRequest("GET", "/some-path", nil)
					recorder := httptest.NewRecorder()
					nextCalled := false

					err := errors.New(faker.Sentence())
					v.On("ValidateRequest", req).Return((*jwt.JSONWebToken)(nil), err)

					next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						nextCalled = true
					})
					mw(next).ServeHTTP(recorder, req)
					if !assert.False(t, nextCalled) {
						return
					}
					v.AssertExpectations(t)

					tst.AssertHTTPErrorResponse(t, tst.NewHTTPErrorPayload(
						http.StatusUnauthorized,
						http.StatusText(http.StatusUnauthorized),
						"Token validation failed",
					), recorder)
				},
			}
		},
		func() testCase {
			v := &mockValidator{}
			return testCase{
				name: "invalid token claims",
				args: args{setup: []MiddlewareOpt{WithValidator(v)}},
				run: func(t *testing.T, mw router.MiddlewareFunc) {

					req := httptest.NewRequest("GET", "/some-path", nil)
					recorder := httptest.NewRecorder()
					nextCalled := false

					err := errors.New(faker.Sentence())
					token := &jwt.JSONWebToken{}
					v.On("ValidateRequest", req).Return(token, nil)
					v.On("Claims", req, token, mock.Anything).Return(err)

					next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						nextCalled = true
					})
					mw(next).ServeHTTP(recorder, req)
					if !assert.False(t, nextCalled) {
						return
					}
					v.AssertExpectations(t)

					tst.AssertHTTPErrorResponse(t, tst.NewHTTPErrorPayload(
						http.StatusUnauthorized,
						http.StatusText(http.StatusUnauthorized),
						"Bad token",
					), recorder)
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

func TestAuthorizeRequest(t *testing.T) {
	type args struct {
		claims *LedgerClaims
	}
	type testCase struct {
		name string
		args args
		run  func(t *testing.T, req *http.Request, recorder *httptest.ResponseRecorder)
	}
	tests := []func() testCase{
		func() testCase {
			return testCase{
				name: "fails if no claims",
				run: func(t *testing.T, req *http.Request, recorder *httptest.ResponseRecorder) {
					nextCalled := false
					next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						nextCalled = true
					})
					AuthorizeRequest(next).ServeHTTP(recorder, req)
					if !assert.False(t, nextCalled) {
						return
					}
					tst.AssertHTTPErrorResponse(t, tst.NewHTTPErrorPayload(
						http.StatusForbidden,
						http.StatusText(http.StatusForbidden),
						"Access token not found",
					), recorder)
				},
			}
		},
		func() testCase {
			return testCase{
				name: "pass if no scope to authorize",
				args: args{claims: &LedgerClaims{}},
				run: func(t *testing.T, req *http.Request, recorder *httptest.ResponseRecorder) {
					nextCalled := false
					nextStatus := rand.Intn(400)
					next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						nextCalled = true
						w.WriteHeader(nextStatus)
					})
					AuthorizeRequest(next).ServeHTTP(recorder, req)
					if !assert.True(t, nextCalled) {
						return
					}
					assert.Equal(t, nextStatus, recorder.Code)
				},
			}
		},
		func() testCase {
			allowedScope := "scope-" + faker.Word()
			actualScope := strings.Join(
				[]string{
					"actual-scope1-" + faker.Word(),
					"actual-scope2-" + faker.Word(),
					allowedScope,
					"actual-scope3-" + faker.Word(),
				},
				" ",
			)
			return testCase{
				name: "pass if allowed scope present",
				args: args{claims: &LedgerClaims{Scope: actualScope}},
				run: func(t *testing.T, req *http.Request, recorder *httptest.ResponseRecorder) {
					nextCalled := false
					nextStatus := rand.Intn(400)
					next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						nextCalled = true
						w.WriteHeader(nextStatus)
					})
					AuthorizeRequest(next, AllowScope(allowedScope)).ServeHTTP(recorder, req)
					if !assert.True(t, nextCalled) {
						return
					}
					assert.Equal(t, nextStatus, recorder.Code)
				},
			}
		},
		func() testCase {
			allowedScope := "allowedScope-" + faker.Word()
			actualScope := strings.Join(
				[]string{
					"actual-scope1-" + faker.Word(),
					"actual-scope2-" + faker.Word(),
					"actual-scope3-" + faker.Word(),
				},
				" ",
			)
			return testCase{
				name: "fail if no allowed scope present",
				args: args{claims: &LedgerClaims{Scope: actualScope}},
				run: func(t *testing.T, req *http.Request, recorder *httptest.ResponseRecorder) {
					nextCalled := false
					next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						nextCalled = true
					})
					AuthorizeRequest(next, AllowScope(allowedScope)).ServeHTTP(recorder, req)
					if !assert.False(t, nextCalled) {
						return
					}
					tst.AssertHTTPErrorResponse(t, tst.NewHTTPErrorPayload(
						http.StatusForbidden,
						http.StatusText(http.StatusForbidden),
						"Missing scope: "+allowedScope,
					), recorder)
				},
			}
		},
	}

	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/some-path", nil)
			recorder := httptest.NewRecorder()
			if tt.args.claims != nil {
				nextContext := ContextWithClaims(req.Context(), tt.args.claims)
				req = req.WithContext(nextContext)
			}
			tt.run(t, req, recorder)
		})
	}
}
