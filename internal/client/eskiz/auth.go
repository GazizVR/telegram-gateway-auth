package eskiz

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// getToken returns the cached token, re-authenticating if it is absent or expired.
func (c *Client) getToken(ctx context.Context) (string, error) {
	c.mu.RLock()
	tok, exp := c.token, c.tokenExp
	c.mu.RUnlock()

	if tok != "" && time.Now().Before(exp) {
		return tok, nil
	}
	return c.login(ctx)
}

// login performs POST /auth/login and caches the returned token.
// Uses double-checked locking so only one goroutine authenticates at a time.
func (c *Client) login(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Another goroutine may have refreshed the token while we waited for the lock.
	if c.token != "" && time.Now().Before(c.tokenExp) {
		return c.token, nil
	}

	form := url.Values{}
	form.Set("email", c.email)
	form.Set("password", c.password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/auth/login", strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("eskiz login: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var resp authResponse
	if err := c.do(req, &resp); err != nil {
		return "", fmt.Errorf("eskiz login: %w", err)
	}
	if resp.Data.Token == "" {
		return "", ErrEmptyToken
	}

	c.token = resp.Data.Token
	c.tokenExp = time.Now().Add(time.Duration(tokenTTL) * time.Second)
	return c.token, nil
}
