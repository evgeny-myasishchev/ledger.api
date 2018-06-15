package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/icrowley/fake"
	. "github.com/smartystreets/goconvey/convey"
	jose "gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
	"ledger.api/pkg/auth"
	"ledger.api/pkg/logging"
)

func TestRouteMiddleware(t *testing.T) {

	Convey("Given router middleware", t, func() {
		app := CreateHTTPApp(HTTPAppConfig{Env: "test"})
		recorder := httptest.NewRecorder()

		handlerCalled := false
		app.RegisterRoutes(func(r *Router) {
			r.handle("GET", "/v1/some-resource", func(req *http.Request, h *HandlerToolkit) (*Response, error) {
				handlerCalled = true
				return h.JSON(JSON{"fake": "string"}), nil
			})
			r.handle("GET", "/v1/should-abort", func(req *http.Request, h *HandlerToolkit) (*Response, error) {
				return h.JSON(JSON{"fake": "string"}), nil
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
					r.handle("GET", "/v1/some-ordered-route", func(req *http.Request, h *HandlerToolkit) (*Response, error) {
						h.Logger.Info("Processing actual route handler")
						So(callCount, ShouldEqual, 3)
						callCount++
						return h.JSON(JSON{"fake": "string"}), nil
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

type jwtTokenSetup struct {
	pwd      string
	iss      string
	aud      string
	rawToken string
	claims   *auth.LedgerClaims
}

func setupJwtToken() (*jwtTokenSetup, error) {
	pwd := fake.SimplePassword()
	iss := fmt.Sprintf("iss-%v", fake.Characters())
	aud := fmt.Sprintf("aud-%v", fake.Characters())
	key := jose.SigningKey{Algorithm: jose.HS256, Key: []byte(pwd)}
	signer, err := jose.NewSigner(key, (&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		return nil, err
	}
	commonClaims := jwt.Claims{
		Issuer:   iss,
		Audience: []string{aud},
		Expiry:   jwt.NewNumericDate(time.Now().Add(time.Minute)),
	}
	claims := auth.LedgerClaims{
		Claims: &commonClaims,
		Scope:  fmt.Sprintf("scope1%v,scope2%v", fake.Characters(), fake.Characters()),
	}
	rawToken, err := jwt.Signed(signer).Claims(claims).CompactSerialize()
	if err != nil {
		return nil, err
	}
	return &jwtTokenSetup{
		pwd:      pwd,
		iss:      iss,
		aud:      aud,
		rawToken: rawToken,
		claims:   &claims,
	}, nil
}

func TestAuthMiddleware(t *testing.T) {
	Convey("Given AuthMiddleware", t, func() {
		tokenSetup, err := setupJwtToken()
		So(err, ShouldBeNil)
		validator := auth.CreateHS256Validator(tokenSetup.pwd, tokenSetup.iss, tokenSetup.aud)
		logger := logging.NewTestLogger()
		initLogger := CreateInitLoggerMiddlewareFunc(logger)
		authMw := CreateAuthMiddlewareFunc(validator)
		middlewareFunc := func(next http.HandlerFunc) http.HandlerFunc {
			return initLogger(authMw((next)))
		}
		req, err := http.NewRequest("GET", "/v1/something", nil)
		if err != nil {
			panic(err)
		}
		Convey("When auth header provided", func() {
			Convey("It should validate the token and set auth data to context", func() {
				recorder := httptest.NewRecorder()
				nextCalled := false
				middleware := middlewareFunc(func(w http.ResponseWriter, req *http.Request) {
					actualClaims := auth.ClaimsFromContext(req.Context())
					So(actualClaims, ShouldResemble, tokenSetup.claims)
					nextCalled = true
				})
				req.Header.Add("Authorization", "Bearer "+tokenSetup.rawToken)
				middleware(recorder, req)
				So(nextCalled, ShouldBeTrue)
			})
			Convey("It should respond with 401 if token validation fails", func() {
				recorder := httptest.NewRecorder()
				nextCalled := false
				middleware := middlewareFunc(func(w http.ResponseWriter, req *http.Request) {
					nextCalled = true
				})
				tokenSetup, err := setupJwtToken()
				So(err, ShouldBeNil)
				req.Header.Add("Authorization", "Bearer "+tokenSetup.rawToken)
				middleware(recorder, req)
				So(nextCalled, ShouldBeFalse)
				So(recorder.Code, ShouldEqual, 401)

				expectedMessage := map[string]interface{}{
					"errors": []interface{}{
						map[string]interface{}{
							"status": strconv.Itoa(http.StatusUnauthorized),
							"title":  http.StatusText(http.StatusUnauthorized),
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

		Convey("When no auth header provided", func() {
			Convey("It should respond with 401 error", func() {
				recorder := httptest.NewRecorder()
				nextCalled := false
				middleware := middlewareFunc(func(w http.ResponseWriter, req *http.Request) {
					nextCalled = true
				})
				middleware(recorder, req)
				So(nextCalled, ShouldBeFalse)
				So(recorder.Code, ShouldEqual, 401)

				expectedMessage := map[string]interface{}{
					"errors": []interface{}{
						map[string]interface{}{
							"status": strconv.Itoa(http.StatusUnauthorized),
							"title":  http.StatusText(http.StatusUnauthorized),
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
