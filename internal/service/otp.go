package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gazizvr/telegram-gateway-auth/internal/domain"
)

type OTPStore interface {
	Get(ctx context.Context, phone string) (*domain.Challenge, error)
	Set(ctx context.Context, phone string, challenge domain.Challenge, ttl time.Duration) error
	Delete(ctx context.Context, phone string) error
}

type OTPProvider interface {
	Send(ctx context.Context, phone, code string, ttl int) (*domain.Challenge, error)
	Verify(ctx context.Context, challenge domain.Challenge, code string) error
}

type OTPService struct {
	provider OTPProvider
	store    OTPStore
}

func NewOTPService(
	provider OTPProvider,
	store OTPStore,
) *OTPService {
	return &OTPService{
		provider: provider,
		store:    store,
	}
}

func (s *OTPService) generateCode() (*string, error) {
	code := ""
	return &code, nil
}

const defaultTTL = 150 * time.Second

func (s *OTPService) Send(ctx context.Context, phone string) error {
	code, err := s.generateCode()
	if err != nil {
		return err
	}
	challenge, err := s.provider.Send(ctx, phone, *code, int(defaultTTL))
	if err != nil {
		return err
	}
	return s.store.Set(ctx, phone, *challenge, defaultTTL)
}

func (s *OTPService) Verify(ctx context.Context, phone, code string) error {
	reqId, err := s.store.Get(ctx, phone)
	if err != nil {
		return err
	}
	if reqId == nil || *reqId == "" {
		return domain.ErrOTPNotFound
	}
	if err := s.provider.Verify(ctx, *reqId, code); err != nil {
		return err
	}
	if err := s.store.Delete(ctx, phone); err != nil {
		return fmt.Errorf("delete challenge failed: %w", err)
	}
	return nil
}
