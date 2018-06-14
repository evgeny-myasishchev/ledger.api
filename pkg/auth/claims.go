package auth

import (
	"context"

	"gopkg.in/square/go-jose.v2/jwt"
)

type contextKeys string

const ledgerClaimsKey contextKeys = "ledgerClaims"

// LedgerClaims - claims structure supported by ledger
type LedgerClaims struct {
	*jwt.Claims
	Scope string `json:"scope"`
}

// ClaimsFromContext returns an instance of LedgerClaims from context
func ClaimsFromContext(ctx context.Context) *LedgerClaims {
	return ctx.Value(ledgerClaimsKey).(*LedgerClaims)
}

// ContextWithClaims creates a new context with LedgerClaims initialized
func ContextWithClaims(parent context.Context, claims *LedgerClaims) context.Context {
	return context.WithValue(parent, ledgerClaimsKey, claims)
}
