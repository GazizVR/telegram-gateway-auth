package service

import (
	"errors"

	telegramgateway "github.com/gazizvr/go-telegram-gateway"
	"github.com/gazizvr/telegram-gateway-auth/internal/domain"
)

func mapGatewayError(err error) error {
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
	case errors.Is(err, telegramgateway.ErrBalanceNotEnough),
		errors.Is(err, telegramgateway.ErrAccessTokenInvalid),
		errors.Is(err, telegramgateway.ErrAccessTokenRequired):
		return domain.ErrGatewayUnavailable
	default:
		return err
	}
}