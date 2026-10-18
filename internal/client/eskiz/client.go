// Package eskiz is a client for the Eskiz SMS gateway API (https://eskiz.uz).
package eskiz

import (
	"net/http"
	"sync"
	"time"
)

// Client is a thread-safe HTTP client for the Eskiz SMS gateway API.
// It lazily authenticates on the first call and caches the Bearer token.
type Client struct {
	email    string
	password string
	from     string
	baseURL  string

	httpClient *http.Client

	mu       sync.RWMutex
	token    string
	tokenExp time.Time
}

// Option is a functional option for configuring a Client.
type Option func(*Client)

// WithHTTPClient replaces the default *http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// WithFrom sets a custom sender alpha-name. Defaults to DefaultFrom ("4546").
func WithFrom(from string) Option {
	return func(c *Client) { c.from = from }
}

// New creates a Client with the given Eskiz credentials.
// Authentication is deferred until the first API call.
func New(email, password string, opts ...Option) *Client {
	c := &Client{
		email:      email,
		password:   password,
		from:       DefaultFrom,
		baseURL:    BaseURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
