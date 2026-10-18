// Package adapter connects the OTP service to external senders
// and translates their errors into domain errors.
package adapter

import (
	"context"
	"time"

	telegramgateway "github.com/gazizvr/go-telegram-gateway"
)

// TelegramSender delivers codes through Telegram Gateway.
type TelegramSender struct {
	client *telegramgateway.Client
}

// NewTelegramSender returns a TelegramSender that uses client.
func NewTelegramSender(client *telegramgateway.Client) *TelegramSender {
	return &TelegramSender{
		client: client,
	}
}

// Send checks that the phone can receive a Telegram message,
// then sends code with the given ttl.
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
