package adapter

import (
	"context"
	"time"

	"github.com/gazizvr/telegram-gateway-auth/internal/domain"
)

// SMSSender is a placeholder fallback used when Eskiz is not configured.
// Send always fails with domain.ErrSenderUnavailable.
type SMSSender struct{}

// NewSMSSender returns a disabled SMSSender.
func NewSMSSender() *SMSSender {
	return &SMSSender{}
}

// Send always returns domain.ErrSenderUnavailable.
func (s *SMSSender) Send(
	ctx context.Context,
	phone string,
	code string,
	ttl time.Duration,
) error {
	return domain.ErrSenderUnavailable
}
