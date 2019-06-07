package auth

import (
	"net/http"

	"github.com/auth0-community/go-auth0"
)

type middlewareCfg struct {
	validator RequestValidator

	whitelistedRoutes map[string]bool
}

// MiddlewareOpt is an auth middleware option
type MiddlewareOpt func(*middlewareCfg)

func withValidator(v RequestValidator) MiddlewareOpt {
	return func(cfg *middlewareCfg) {
		cfg.validator = v
	}
}

// NewMiddleware creates auth middleware that will validate token presense and token validity
func NewMiddleware(setup ...MiddlewareOpt) func(next http.HandlerFunc) http.HandlerFunc {
	cfg := middlewareCfg{}
	for _, opt := range setup {
		opt(&cfg)
	}
	validator := cfg.validator
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			token, err := validator.ValidateRequest(req)
			if err != nil {
				if err == auth0.ErrTokenNotFound {
					logger.Debug(req.Context(), "No token found")
					next(w, req)
					return
				}
				logger.WithError(err).Error(req.Context(), "Token validation failed")
				// TODO: Respond with error
				// TODO: Ignore token not found error
				return
			}
			claims := LedgerClaims{}
			if err := validator.Claims(req, token, &claims); err != nil {
				//TODO: Respond with error
				logger.WithError(err).Error(req.Context(), "Failed to get claims")
				return
			}

			nextContext := ContextWithClaims(req.Context(), &claims)
			nextReq := req.WithContext(nextContext)

			next(w, nextReq)
		}
	}
}

// CreateAuthMiddlewareFunc returns auth middleware func that creates auth middleware
// TODO: Port this stuff
// func CreateAuthMiddlewareFunc(params AuthMiddlewareParams) RouterMiddlewareFunc {
// 	validator := params.Validator
// 	whitelistedRoutes := params.WhitelistedRoutes
// 	shouldWhitelist := func(err error, req *http.Request) bool {
// 		if whitelistedRoutes == nil {
// 			return false
// 		}
// 		if err != auth0.ErrTokenNotFound {
// 			return false
// 		}
// 		path := strings.TrimRight(req.URL.Path, "/")
// 		return whitelistedRoutes[path]
// 	}

// 	return func(next http.HandlerFunc) http.HandlerFunc {
// 		return func(w http.ResponseWriter, req *http.Request) {
// 			logger := logging.FromContext(req.Context())
// 			token, err := validator.ValidateRequest(req)
// 			if err != nil {
// 				if shouldWhitelist(err, req) {
// 					next(w, req)
// 					return
// 				}
// 				logger.WithError(err).Error("Token validation failed")
// 				respondWithErrorStatus(w, http.StatusUnauthorized)
// 				return
// 			}

// 			claims := auth.LedgerClaims{}

// 			validator.Claims(req, token, &claims)
// 			if err != nil {
// 				logger.WithError(err).Error("Failed to get claims")
// 				respondWithErrorStatus(w, http.StatusUnauthorized)
// 				return
// 			}

// 			nextContext := auth.ContextWithClaims(req.Context(), &claims)
// 			nextReq := req.WithContext(nextContext)

// 			next(w, nextReq)
// 		}
// 	}
// }

// // RequireScopes action handler middleware wrapper that will
// // verify if scope claim of a token includes scopes provided
// func RequireScopes(handler HandlerFunc, scopes ...string) HandlerFunc {
// 	return HandlerFunc(func(req *http.Request, h *HandlerToolkit) (*Response, error) {
// 		claims := auth.ClaimsFromContext(req.Context())
// 		if claims == nil {
// 			h.Logger.Info("Request has not been initialized with claims, responding with 404")
// 			return nil, HTTPError{
// 				Status: http.StatusNotFound,
// 				Errors: []*jsonapi.ErrorObject{
// 					{
// 						Status: strconv.Itoa(http.StatusNotFound),
// 						Title:  http.StatusText(http.StatusNotFound),
// 					},
// 				},
// 			}
// 		}

// 		authorizedScopes := make(map[string]bool)
// 		for _, authorizedScope := range strings.Split(claims.Scope, " ") {
// 			authorizedScopes[authorizedScope] = true
// 		}

// 		var missingScopes []string
// 		for _, requiredScope := range scopes {
// 			if !authorizedScopes[requiredScope] {
// 				missingScopes = append(missingScopes, requiredScope)
// 			}
// 		}

// 		if len(missingScopes) > 0 {
// 			h.Logger.Info("Failed to authorize request. Missing scopes: %v", missingScopes)
// 			return nil, HTTPError{
// 				Status: http.StatusForbidden,
// 				Errors: []*jsonapi.ErrorObject{
// 					{
// 						Status: strconv.Itoa(http.StatusForbidden),
// 						Title:  http.StatusText(http.StatusForbidden),
// 						Detail: fmt.Sprintf("Missing scopes: %v", missingScopes),
// 					},
// 				},
// 			}
// 		}

// 		return handler(req, h)
// 	})
// }
