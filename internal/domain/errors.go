package domain

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidPhoneNumber  = errors.New("invalid phone number")
	ErrOTPNotFound         = errors.New("otp session not found or expired")
	ErrInvalidCode         = errors.New("invalid verification code")
	ErrCodeExpired         = errors.New("verification code expired")
	ErrMaxAttemptsExceeded = errors.New("maximum verification attempts exceeded")
	ErrMessageAlreadySent  = errors.New("verification message already sent")
	ErrTooManyRequests     = errors.New("too many requests")
	ErrGatewayUnavailable  = errors.New("telegram gateway service unavailable")
)

type FloodWaitError struct {
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
