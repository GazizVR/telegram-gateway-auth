package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gazizvr/telegram-gateway-auth/internal/domain"
)

func handleError(ctx context.Context, w http.ResponseWriter, err error) {
	slog.ErrorContext(ctx, err.Error())
	var floodErr *domain.FloodWaitError
	if errors.As(err, &floodErr) {
		seconds := 0
		if floodErr.WaitTime > 0 {
			seconds = int(floodErr.WaitTime.Seconds())
			w.Header().Set("Retry-After", strconv.Itoa(seconds))
		}
		writeErrorWithRetry(w, http.StatusTooManyRequests, "FLOOD_WAIT", floodErr.Error(), seconds)
		return
	}

	switch {
	case errors.Is(err, domain.ErrInvalidPhoneNumber):
		writeError(w, http.StatusBadRequest, "INVALID_PHONE_NUMBER", err.Error())
	case errors.Is(err, domain.ErrInvalidCode):
		writeError(w, http.StatusBadRequest, "INVALID_CODE", err.Error())
	case errors.Is(err, domain.ErrCodeExpired):
		writeError(w, http.StatusBadRequest, "CODE_EXPIRED", err.Error())
	case errors.Is(err, domain.ErrMaxAttemptsExceeded):
		writeError(w, http.StatusTooManyRequests, "MAX_ATTEMPTS_EXCEEDED", err.Error())
	case errors.Is(err, domain.ErrTooManyRequests):
		writeError(w, http.StatusTooManyRequests, "TOO_MANY_REQUESTS", err.Error())
	case errors.Is(err, domain.ErrMessageAlreadySent):
		writeError(w, http.StatusConflict, "MESSAGE_ALREADY_SENT", err.Error())
	case errors.Is(err, domain.ErrOTPNotFound):
		writeError(w, http.StatusNotFound, "OTP_NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrPhoneNotReachable):
		writeError(w, http.StatusUnprocessableEntity, "PHONE_NOT_REACHABLE", err.Error())
	case errors.Is(err, domain.ErrSenderUnavailable):
		writeError(w, http.StatusBadGateway, "SENDER_UNAVAILABLE", "sender service unavailable")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}
