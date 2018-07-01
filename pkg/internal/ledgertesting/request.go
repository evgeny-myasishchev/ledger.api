package ledgertesting

import (
	"io"
	"net/http"

	"ledger.api/pkg/auth"
)

type requestOptions struct {
	body  io.Reader
	scope string
}

// RequestOption function to set request options
type RequestOption func(opts *requestOptions)

// WithScopeClaim will set a scope claim to initialize request with
func WithScopeClaim(scope string) RequestOption {
	return func(opts *requestOptions) {
		opts.scope = scope
	}
}

// NewRequest creates a new instance of the http request for testing purposes
func NewRequest(method string, url string, opts ...RequestOption) *http.Request {
	reqOpts := requestOptions{}
	for _, opt := range opts {
		opt(&reqOpts)
	}

	req, err := http.NewRequest(method, url, reqOpts.body)
	if err != nil {
		panic(err)
	}

	if reqOpts.scope != "" {
		req = req.WithContext(
			auth.ContextWithClaims(
				req.Context(),
				&auth.LedgerClaims{Scope: "read:ledgers write:ledgers"},
			))
	}

	return req
}
