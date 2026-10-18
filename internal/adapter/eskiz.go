package adapter

import (
	"context"
	"log/slog"
	"time"

	"github.com/gazizvr/telegram-gateway-auth/internal/client/eskiz"
)

type EskizSender struct {
	client *eskiz.Client
}

func NewEskizSender(client *eskiz.Client) *EskizSender {
	return &EskizSender{
		client: client,
	}
}

func (s *EskizSender) Send(
	ctx context.Context,
	phone string,
	code string,
	ttl time.Duration,
) error {
	if err := s.client.Send(ctx, phone, code, ttl); err != nil {
		slog.ErrorContext(ctx, "eskiz: failed to send sms", "error", err, "phone", phone)
		return mapEskizError(err)
	}
	return nil
}
