package service

import (
	"context"
	"errors"
	"time"

	telegramgateway "github.com/gazizvr/go-telegram-gateway"
	"github.com/gazizvr/telegram-gateway-auth/internal/domain"
)

type OTPStore interface {
	Get(ctx context.Context, phone string) (*string, error)
	Set(ctx context.Context, phone, reqID string, expAt time.Duration) error
	Delete(ctx context.Context, phone string) error
}

type client interface {
	SendVerificationMessage(
		ctx context.Context,
		req telegramgateway.SendVerificationMessageRequest,
	) (*telegramgateway.RequestStatus, error)
	CheckSendAbility(
		ctx context.Context,
		req telegramgateway.CheckSendAbilityRequest,
	) (*telegramgateway.RequestStatus, error)
	CheckVerificationStatus(
		ctx context.Context,
		req telegramgateway.CheckVerificationStatusRequest,
	) (*telegramgateway.RequestStatus, error)
	RevokeVerificationMessage(
		ctx context.Context,
		req telegramgateway.RevokeVerificationMessageRequest,
	) (bool, error)
}

type OTPService struct {
	cl    client
	store OTPStore
}

func NewOTPService(
	client client,
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
		return mapGatewayError(err)
	}
	req := telegramgateway.SendVerificationMessageRequest{
		PhoneNumber: resp.PhoneNumber,
		RequestID:   resp.RequestID,
		CodeLength:  defaultCodeLength,
	}
	if _, err := s.cl.SendVerificationMessage(ctx, req); err != nil {
		return mapGatewayError(err)
	}
	return s.store.Set(ctx, phone, req.RequestID, codeDefaultTTL)
}

func (s *OTPService) Verify(ctx context.Context, phone, code string) error {
	reqId, err := s.store.Get(ctx, phone)
	if err != nil {
		return err
	}
	if reqId == nil || *reqId == "" {
		return domain.ErrOTPNotFound
	}
	req := telegramgateway.CheckVerificationStatusRequest{
		RequestID: *reqId,
		Code:      code,
	}
	resp, err := s.cl.CheckVerificationStatus(ctx, req)
	if err != nil {
		return mapGatewayError(err)
	}
	if resp.VerificationStatus == nil {
		return domain.ErrInvalidCode
	}
	switch resp.VerificationStatus.Status {
	case telegramgateway.VerificationCodeValid:
		_ = s.store.Delete(ctx, phone)
		_, _ = s.cl.RevokeVerificationMessage(
			ctx,
			telegramgateway.RevokeVerificationMessageRequest{RequestID: *reqId},
		)
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

func mapGatewayError(err error) error {
	if err == nil {
		return nil
	}
	var floodErr *telegramgateway.FloodError
	if errors.As(err, &floodErr) {
		return &domain.FloodWaitError{WaitTime: floodErr.WaitTime}
	}
	switch {
	case errors.Is(err, telegramgateway.ErrPhoneNumberInvalid),
		errors.Is(err, telegramgateway.ErrPhoneNumberMismatch):
		return domain.ErrInvalidPhoneNumber
	case errors.Is(err, telegramgateway.ErrMessageAlreadySent):
		return domain.ErrMessageAlreadySent
	case errors.Is(err, telegramgateway.ErrCodeInvalid):
		return domain.ErrInvalidCode
	case errors.Is(err, telegramgateway.ErrCodeExpired):
		return domain.ErrCodeExpired
	case errors.Is(err, telegramgateway.ErrCodeMaxAttemptsExceeded):
		return domain.ErrMaxAttemptsExceeded
	case errors.Is(err, telegramgateway.ErrRequestIDInvalid),
		errors.Is(err, telegramgateway.ErrRequestIDRequired):
		return domain.ErrOTPNotFound
	case errors.Is(err, telegramgateway.ErrBalanceNotEnough),
		errors.Is(err, telegramgateway.ErrAccessTokenInvalid),
		errors.Is(err, telegramgateway.ErrAccessTokenRequired):
		return domain.ErrGatewayUnavailable
	default:
		return err
	}
}
