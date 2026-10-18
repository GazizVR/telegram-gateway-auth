package eskiz

import "errors"

const (
	// BaseURL is the root endpoint of the Eskiz SMS gateway API.
	BaseURL = "https://notify.eskiz.uz/api"

	// DefaultFrom is the sender ID used when no custom alpha-name is set.
	DefaultFrom = "4546"

	// tokenTTL is how long a cached token is considered valid.
	// Eskiz tokens last 30 days; we subtract 1 hour to refresh before expiry.
	tokenTTL = 29*24*60*60 + 23*60*60 // seconds
)

// Errors returned by Client methods.
var (
	ErrUnauthorized = errors.New("eskiz: unauthorized – token is invalid or expired")
	ErrEmptyToken   = errors.New("eskiz: received empty token from server")
	ErrBadRequest   = errors.New("eskiz: bad request")
	ErrServerError  = errors.New("eskiz: internal server error")
)
