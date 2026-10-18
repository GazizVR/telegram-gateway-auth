package adapter

import (
	"errors"

	telegramgateway "github.com/gazizvr/go-telegram-gateway"
	"github.com/gazizvr/telegram-gateway-auth/internal/client/eskiz"
	"github.com/gazizvr/telegram-gateway-auth/internal/domain"
)

func mapTelegramError(err error) error {
	if err == nil {
		return nil
	}
	var floodErr *telegramgateway.FloodError
	if errors.As(err, &floodErr) {
		return &domain.FloodWaitError{WaitTime: floodErr.WaitTime}
	}
	switch {
	case errors.Is(err, telegramgateway.ErrPhoneNumberInvalid),
		errors.Is(err, telegramgateway.ErrPhoneNumberMismatch):
		return domain.ErrInvalidPhoneNumber
	case errors.Is(err, telegramgateway.ErrMessageAlreadySent):
		return domain.ErrMessageAlreadySent
	case errors.Is(err, telegramgateway.ErrCodeInvalid):
		return domain.ErrInvalidCode
	case errors.Is(err, telegramgateway.ErrCodeExpired):
		return domain.ErrCodeExpired
	case errors.Is(err, telegramgateway.ErrCodeMaxAttemptsExceeded):
		return domain.ErrMaxAttemptsExceeded
	case errors.Is(err, telegramgateway.ErrRequestIDInvalid),
		errors.Is(err, telegramgateway.ErrRequestIDRequired):
		return domain.ErrOTPNotFound
	case errors.Is(err, telegramgateway.ErrPhoneNumberNotFound),
		errors.Is(err, telegramgateway.ErrPhoneNumberNotAvailable):
		return domain.ErrPhoneNotReachable
	case errors.Is(err, telegramgateway.ErrBalanceNotEnough),
		errors.Is(err, telegramgateway.ErrAccessTokenInvalid),
		errors.Is(err, telegramgateway.ErrAccessTokenRequired):
		return domain.ErrSenderUnavailable
	default:
		return err
	}
}

// mapEskizError translates eskiz client errors into domain errors
// that the OTPService fallback logic understands.
func mapEskizError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	// Auth / infrastructure problems — the SMS provider itself is down.
	case errors.Is(err, eskiz.ErrUnauthorized),
		errors.Is(err, eskiz.ErrEmptyToken),
		errors.Is(err, eskiz.ErrServerError):
		return domain.ErrSenderUnavailable

	// The phone number was rejected by the API (wrong format, not a UZ number, etc.).
	case errors.Is(err, eskiz.ErrBadRequest):
		return domain.ErrPhoneNotReachable

	default:
		return domain.ErrSenderUnavailable
	}
}
