package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gazizvr/telegram-gateway-auth/internal/domain"
)

type OTPService interface {
	Send(ctx context.Context, phone string) error
	Verify(ctx context.Context, phone, code string) error
}

type OTPHandler struct {
	src OTPService
}

func NewOTPHandler(service OTPService) *OTPHandler {
	return &OTPHandler{
		src: service,
	}
}

type sendRequest struct {
	Phone string `json:"phone"`
}

type sendResponse struct {
	OK bool `json:"ok"`
}

func (h *OTPHandler) Send(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req sendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}
	defer r.Body.Close()

	ctx := r.Context()
	if err := h.src.Send(ctx, req.Phone); err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, &sendResponse{OK: true})
}

type verifyRequest struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

type verifyResponse struct {
	OK bool `json:"ok"`
}

func (h *OTPHandler) Verify(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req verifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}
	defer r.Body.Close()

	ctx := r.Context()
	if err := h.src.Verify(ctx, req.Phone, req.Code); err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, &verifyResponse{OK: true})
}

func (h *OTPHandler) handleError(w http.ResponseWriter, err error) {
	var floodErr *domain.FloodWaitError
	if errors.As(err, &floodErr) {
		if floodErr.WaitTime > 0 {
			w.Header().Set("Retry-After", strconv.Itoa(int(floodErr.WaitTime.Seconds())))
		}
		writeError(w, http.StatusTooManyRequests, "TOO_MANY_REQUESTS", floodErr.Error())
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
	case errors.Is(err, domain.ErrGatewayUnavailable):
		writeError(w, http.StatusBadGateway, "GATEWAY_UNAVAILABLE", "telegram gateway service unavailable")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}
