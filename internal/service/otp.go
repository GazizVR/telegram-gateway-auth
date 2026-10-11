package service

import (
	"context"
	"time"

	telegramgateway "github.com/gazizvr/go-telegram-gateway"
)

type OTPStore interface {
	Get(ctx context.Context, phone string) (*string, error)
	Set(ctx context.Context, phone, reqID string, expAt time.Duration) error
	Delete(ctx context.Context, phone string) error
}

type OTPService struct {
	cl    telegramgateway.Client
	store OTPStore
}

func NewOTPService(
	client telegramgateway.Client,
	store OTPStore,
) *OTPService {
	return &OTPService{
		cl:    client,
		store: store,
	}
}

const defaultCodeLength int = 6
const codeDefaultTTL = 150 * time.Second

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
	return s.store.Set(ctx, phone, req.RequestID, codeDefaultTTL)
}

func (s *OTPService) Verify(ctx context.Context, phone, code string) error {
	reqId, err := s.store.Get(ctx, phone)
	if err != nil {
		return err
	}
	req := telegramgateway.CheckVerificationStatusRequest{
		RequestID: *reqId,
		Code:      code,
	}
	resp, err := s.cl.CheckVerificationStatus(ctx, req)
	if err != nil {
		return err
	}
	if resp.VerificationStatus.Status == telegramgateway.VerificationCodeValid {
		s.store.Delete(ctx, phone)
		s.cl.RevokeVerificationMessage(
			ctx,
			telegramgateway.RevokeVerificationMessageRequest{RequestID: *reqId},
		)
	} else {
		return ErrCodeInvalid
	}
	return nil
}
