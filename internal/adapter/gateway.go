package adapter

import (
	"context"
	"time"

	telegramgateway "github.com/gazizvr/go-telegram-gateway"
)

type TelegramGateway struct {
	client *telegramgateway.Client
}

func NewTelegramGateway(client *telegramgateway.Client) *TelegramGateway {
	return &TelegramGateway{
		client: client,
	}
}

func (c *TelegramGateway) Send(
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
		return mapGatewayError(err)
	}
	ttlSec := int(ttl.Seconds())
	req := telegramgateway.SendVerificationMessageRequest{
		PhoneNumber: resp.PhoneNumber,
		RequestID:   resp.RequestID,
		Code:        code,
		TTL:         ttlSec,
	}
	if _, err := c.client.SendVerificationMessage(ctx, req); err != nil {
		return mapGatewayError(err)
	}
	return nil
}