package adapter

import (
	"context"
	"time"

	"github.com/gazizvr/telegram-gateway-auth/internal/domain"
)

type SMSSender struct{}

func NewSMSSender() *SMSSender {
	return &SMSSender{}
}

func (s *SMSSender) Send(
	ctx context.Context,
	phone string,
	code string,
	ttl time.Duration,
) error {
	return domain.ErrSenderUnavailable
}
