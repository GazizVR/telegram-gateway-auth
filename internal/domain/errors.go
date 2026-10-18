// Package domain defines the errors shared across the service layers.
package domain

import (
	"errors"
	"fmt"
	"time"
)

// Sentinel errors. Senders and stores return them, the service reacts to them,
// and the HTTP handler maps them to response codes.
var (
	ErrInvalidPhoneNumber  = errors.New("invalid phone number")
	ErrOTPNotFound         = errors.New("otp session not found or expired")
	ErrInvalidCode         = errors.New("invalid verification code")
	ErrCodeExpired         = errors.New("verification code expired")
	ErrMaxAttemptsExceeded = errors.New("maximum verification attempts exceeded")
	ErrMessageAlreadySent  = errors.New("verification message already sent")
	ErrTooManyRequests     = errors.New("too many requests")
	ErrSenderUnavailable   = errors.New("sender service unavailable")
	ErrPhoneNotReachable   = errors.New("phone number not reachable")
)

// FloodWaitError reports that sending is rate limited.
// It unwraps to ErrTooManyRequests.
type FloodWaitError struct {
	// WaitTime is how long to wait before retrying. Zero if unknown.
	WaitTime time.Duration
}

func (e *FloodWaitError) Error() string {
	if e.WaitTime > 0 {
		return fmt.Sprintf("flood wait: retry after %s", e.WaitTime)
	}
	return "flood wait: rate limit exceeded"
}

func (e *FloodWaitError) Unwrap() error {
	return ErrTooManyRequests
}
