package auth

import (
	"net/http"
	"strings"

	"ledger.api/pkg/core/router"

	"github.com/auth0-community/go-auth0"
)

type middlewareCfg struct {
	validator RequestValidator

	whitelistedRoutes map[string]bool
}

// MiddlewareOpt is an auth middleware option
type MiddlewareOpt func(*middlewareCfg)

// WithValidator - setup request validator
func WithValidator(v RequestValidator) MiddlewareOpt {
	return func(cfg *middlewareCfg) {
		cfg.validator = v
	}
}

// WithAuth0Validator - setup validator instance configured to use
// jwt tokens issued by auth0
func WithAuth0Validator(iss string, aud string) MiddlewareOpt {
	return func(cfg *middlewareCfg) {
		cfg.validator = createAuth0Validator(iss, aud)
	}
}

// NewMiddleware creates auth middleware that will validate token if present and inject claims into a context
func NewMiddleware(setup ...MiddlewareOpt) router.MiddlewareFunc {
	cfg := middlewareCfg{}
	for _, opt := range setup {
		opt(&cfg)
	}
	validator := cfg.validator
	return func(next http.Handler) http.Handler {
		respondUnauthorized := func(w http.ResponseWriter, message string) {
			errorResponse := &router.HTTPError{
				StatusCode: http.StatusUnauthorized,
				Status:     http.StatusText(http.StatusUnauthorized),
				Message:    message,
			}
			errorResponse.Send(w)
		}
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			token, err := validator.ValidateRequest(req)
			if err != nil {
				// Not failing if no token found
				// Authorization is a subject of another middleware
				if err == auth0.ErrTokenNotFound {
					logger.Debug(req.Context(), "No token found")
					next.ServeHTTP(w, req)
					return
				}
				logger.WithError(err).Error(req.Context(), "Token validation failed")
				respondUnauthorized(w, "Token validation failed")
				return
			}
			claims := LedgerClaims{}
			if err := validator.Claims(req, token, &claims); err != nil {
				logger.WithError(err).Error(req.Context(), "Failed to get claims")
				respondUnauthorized(w, "Bad token")
				return
			}

			nextContext := ContextWithClaims(req.Context(), &claims)
			nextReq := req.WithContext(nextContext)

			next.ServeHTTP(w, nextReq)
		})
	}
}

type authorizeRequestCfg struct {
	allowedScope []string
}

// AuthorizeOpt is an option for AuthorizeRequest wrapper
type AuthorizeOpt func(*authorizeRequestCfg)

// AllowScope is an option to allow particular scope
func AllowScope(scope ...string) AuthorizeOpt {
	return func(cfg *authorizeRequestCfg) {
		cfg.allowedScope = append(cfg.allowedScope, scope...)
	}
}

// AuthorizeRequest is a request handler wrapper to validate token presence
// and optionally validate if token includes particular scope
func AuthorizeRequest(next http.Handler, setup ...AuthorizeOpt) http.Handler {
	cfg := &authorizeRequestCfg{}
	for _, opt := range setup {
		opt(cfg)
	}

	respondForbidden := func(w http.ResponseWriter, message string) {
		errorResponse := &router.HTTPError{
			StatusCode: http.StatusForbidden,
			Status:     http.StatusText(http.StatusForbidden),
			Message:    message,
		}
		errorResponse.Send(w)
	}

	toStringSet := func(strings []string) map[string]bool {
		result := make(map[string]bool, len(strings))
		for _, str := range strings {
			result[str] = true
		}
		return result
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		claims := ClaimsFromContext(req.Context())
		if claims == nil {
			logger.Info(req.Context(), "Failed to authorize request. No claims present")
			respondForbidden(w, "Access token not found")
			return
		}
		scopeTokens := toStringSet(strings.FieldsFunc(claims.Scope, func(s rune) bool {
			return s == ' '
		}))
		for _, allowedScope := range cfg.allowedScope {
			if ok := scopeTokens[allowedScope]; !ok {
				logger.Info(req.Context(), "Failed to authorize request. Missing scopes: %v", allowedScope)
				respondForbidden(w, "Missing scope: "+allowedScope)
				return
			}
		}
		next.ServeHTTP(w, req)
	})
}
