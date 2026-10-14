package adapter

import (
	"context"
	"time"

	telegramgateway "github.com/gazizvr/go-telegram-gateway"
)

type TelegramSender struct {
	client *telegramgateway.Client
}

func NewTelegramSender(client *telegramgateway.Client) *TelegramSender {
	return &TelegramSender{
		client: client,
	}
}

func (c *TelegramSender) Send(
	ctx context.Context,
	phone string,
	code string,
	ttl time.Duration,
) error {
	resp, err := c.client.CheckSendAbility(
		ctx,
		telegramgateway.CheckSendAbilityRequest{
			PhoneNumber: phone,
		},
	)
	if err != nil {
		return mapTelegramError(err)
	}
	ttlSec := int(ttl.Seconds())
	req := telegramgateway.SendVerificationMessageRequest{
		PhoneNumber: resp.PhoneNumber,
		RequestID:   resp.RequestID,
		Code:        code,
		TTL:         ttlSec,
	}
	if _, err := c.client.SendVerificationMessage(ctx, req); err != nil {
		return mapTelegramError(err)
	}
	return nil
}
