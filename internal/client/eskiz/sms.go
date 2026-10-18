package eskiz

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SendSMS sends a single SMS and returns the message ID assigned by Eskiz.
// phone must be in international format without '+', e.g. "998901234567".
// On a 401 response the client re-authenticates once and retries automatically.
func (c *Client) SendSMS(ctx context.Context, phone, message string) (string, error) {
	tok, err := c.getToken(ctx)
	if err != nil {
		return "", err
	}

	form := url.Values{}
	form.Set("mobile_phone", phone)
	form.Set("message", message)
	form.Set("from", c.from)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/message/sms/send", strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("eskiz send: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+tok)

	var resp sendResponse
	if err := c.do(req, &resp); err != nil {
		if errors.Is(err, ErrUnauthorized) {
			if _, loginErr := c.login(ctx); loginErr != nil {
				return "", loginErr
			}
			return c.SendSMS(ctx, phone, message)
		}
		return "", fmt.Errorf("eskiz send: %w", err)
	}
	return resp.ID, nil
}

// Send implements the service.OTPSender interface.
// It formats the OTP text with the remaining TTL and delegates to SendSMS.
func (c *Client) Send(ctx context.Context, phone, code string, ttl time.Duration) error {
	minutes := int(ttl.Minutes())
	if minutes < 1 {
		minutes = 1
	}
	msg := fmt.Sprintf("Your verification code: %s. Valid for %d min.", code, minutes)
	_, err := c.SendSMS(ctx, phone, msg)
	return err
}
