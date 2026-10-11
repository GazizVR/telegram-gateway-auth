package service

import (
	"context"

	telegramgateway "github.com/gazizvr/go-telegram-gateway"
)

type OTPService struct {
	cl telegramgateway.Client
}

const defaultCodeLength int = 6

func (s *OTPService) Send(ctx context.Context, phone string) error {
	resp, err := s.cl.CheckSendAbility(
		ctx,
		telegramgateway.CheckSendAbilityRequest{
			PhoneNumber: phone,
		},
	)
	if err != nil {
		return err
	}
	req := telegramgateway.SendVerificationMessageRequest{
		PhoneNumber: resp.PhoneNumber,
		RequestID:   resp.RequestID,
		CodeLength:  defaultCodeLength,
	}
	if _, err := s.cl.SendVerificationMessage(ctx, req); err != nil {
		return err
	}
	return nil
}
