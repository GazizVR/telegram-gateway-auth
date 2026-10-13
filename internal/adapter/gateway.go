package adapter

import (
	"context"
	"time"

	telegramgateway "github.com/gazizvr/go-telegram-gateway"
	"github.com/gazizvr/telegram-gateway-auth/internal/domain"
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
) (*domain.Challenge, error) {
	resp, err := c.client.CheckSendAbility(
		ctx,
		telegramgateway.CheckSendAbilityRequest{
			PhoneNumber: phone,
		},
	)
	if err != nil {
		return nil, mapGatewayError(err)
	}
	ttlSec := int(ttl.Seconds())
	req := telegramgateway.SendVerificationMessageRequest{
		PhoneNumber: resp.PhoneNumber,
		RequestID:   resp.RequestID,
		Code:        code,
		TTL:         ttlSec,
	}
	if _, err := c.client.SendVerificationMessage(ctx, req); err != nil {
		return nil, mapGatewayError(err)
	}
	challenge := domain.Challenge(resp.RequestID)
	return &challenge, nil
}

func (c *TelegramGateway) Verify(
	ctx context.Context,
	challenge domain.Challenge,
	code string,
) error {
	req := telegramgateway.CheckVerificationStatusRequest{
		RequestID: string(challenge),
		Code:      code,
	}
	resp, err := c.client.CheckVerificationStatus(ctx, req)
	if err != nil {
		return mapGatewayError(err)
	}
	if resp.VerificationStatus == nil {
		return domain.ErrInvalidCode
	}
	switch resp.VerificationStatus.Status {
	case telegramgateway.VerificationCodeValid:
		return nil
	case telegramgateway.VerificationCodeInvalid:
		return domain.ErrInvalidCode
	case telegramgateway.VerificationExpired:
		return domain.ErrCodeExpired
	case telegramgateway.VerificationMaxAttempts:
		return domain.ErrMaxAttemptsExceeded
	default:
		return domain.ErrInvalidCode
	}
}
