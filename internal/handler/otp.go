package handler

import (
	"context"
	"encoding/json"
	"net/http"
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
		handleError(w, err)
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
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, &verifyResponse{OK: true})
}
