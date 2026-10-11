package handler

import (
	"context"
	"encoding/json"
	"net/http"
)

type OTPService interface {
	Send(ctx context.Context, phone string) error
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
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	defer r.Body.Close()

	ctx := r.Context()
	if err := h.src.Send(ctx, req.Phone); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, &sendResponse{OK: true})
}
