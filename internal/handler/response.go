package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func writeJSON(
	w http.ResponseWriter,
	status int,
	v any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write respose", "err", err)
	}
}

type errorResponse struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	RetryAfter int    `json:"retry_after,omitempty"`
}

func writeError(
	w http.ResponseWriter,
	status int,
	code, msg string,
) {
	writeJSON(w, status, errorResponse{Code: code, Message: msg})
}

func writeErrorWithRetry(
	w http.ResponseWriter,
	status int,
	code, msg string,
	retryAfter int,
) {
	writeJSON(w, status, errorResponse{Code: code, Message: msg, RetryAfter: retryAfter})
}
