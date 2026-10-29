// Package handler exposes the OTP service over HTTP.
package handler

import (
	"context"
	"encoding/json"
	"net/http"
)

// OTPService is the business logic used by OTPHandler.
type OTPService interface {
	Send(ctx context.Context, phone string) error
	Verify(ctx context.Context, phone, code string) error
}

// OTPHandler serves the send-code and verify-code endpoints.
type OTPHandler struct {
	src OTPService
}

// NewOTPHandler returns an OTPHandler backed by service.
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

// Send handles POST /v1/send-code with body {"phone": "..."}.
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
		handleError(r.Context(), w, err)
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

// Verify handles POST /v1/verify-code with body {"phone": "...", "code": "..."}.
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
		handleError(r.Context(), w, err)
		return
	}

	writeJSON(w, http.StatusOK, &verifyResponse{OK: true})
}
