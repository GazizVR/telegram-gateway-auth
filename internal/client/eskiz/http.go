package eskiz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// errorBody is used only to detect token-expiry messages inside 500 responses.
type errorBody struct {
	Message string `json:"message"`
}

// do executes req, reads the full response body, and decodes JSON into v.
// It maps well-known HTTP status codes to typed sentinel errors so callers
// can use errors.Is for branching.
//
// NOTE: Eskiz returns HTTP 500 with {"message":"Token has expired"} when the
// token expires, instead of a proper 401. We detect that case here and return
// ErrUnauthorized so the retry logic in SendSMS triggers re-authentication.
func (c *Client) do(req *http.Request, v any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		// success – fall through to decode
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return fmt.Errorf("%w: %s", ErrBadRequest, body)
	default:
		if resp.StatusCode >= 500 {
			// Eskiz sends 500 + "Token has expired" instead of 401.
			if isTokenExpired(body) {
				return ErrUnauthorized
			}
			return fmt.Errorf("%w: status %d: %s", ErrServerError, resp.StatusCode, body)
		}
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
	}

	if v != nil {
		if err := json.Unmarshal(body, v); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// isTokenExpired reports whether the response body contains Eskiz's
// token-expiry message returned with a 500 status code.
func isTokenExpired(body []byte) bool {
	if !bytes.Contains(body, []byte("expired")) {
		return false
	}
	var e errorBody
	_ = json.Unmarshal(body, &e)
	return e.Message == "Token has expired"
}
