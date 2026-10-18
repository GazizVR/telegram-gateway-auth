package eskiz

type tokenData struct {
	Token string `json:"token"`
}

type authResponse struct {
	Message   string    `json:"message"`
	Data      tokenData `json:"data"`
	TokenType string    `json:"token_type"`
}

type sendResponse struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}
