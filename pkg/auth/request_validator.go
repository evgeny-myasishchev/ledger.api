package auth

import (
	"net/http"

	auth0 "github.com/auth0-community/go-auth0"
	jose "gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

// RequestValidator will validate request and return error or JWT token
type RequestValidator interface {
	ValidateRequest(r *http.Request) (*jwt.JSONWebToken, error)
	Claims(r *http.Request, token *jwt.JSONWebToken, values ...interface{}) error
}

// CreateAuth0Validator returns validator instance configured to use
// jwt tokens issued by auth0
func CreateAuth0Validator(iss string, aud string) RequestValidator {
	client := auth0.NewJWKClient(auth0.JWKClientOptions{URI: iss + ".well-known/jwks.json"}, nil)
	configuration := auth0.NewConfiguration(client, []string{aud}, iss, jose.RS256)
	validator := auth0.NewValidator(configuration, nil)
	return validator
}

// CreateHS256Validator returns validator instance configured to use
// jwt tokens signed using HS256 alg
func CreateHS256Validator(secret string, iss string, aud string) RequestValidator {
	key := jose.SigningKey{Algorithm: jose.HS256, Key: []byte(secret)}
	keyProvider := auth0.NewKeyProvider(key)
	configuration := auth0.NewConfiguration(keyProvider, []string{aud}, iss, jose.HS256)
	validator := auth0.NewValidator(configuration, nil)
	return validator
}
