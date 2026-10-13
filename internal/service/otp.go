package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/gazizvr/telegram-gateway-auth/internal/domain"
)

type OTPStore interface {
	Get(ctx context.Context, phone string) (*domain.Challenge, error)
	Set(ctx context.Context, phone string, challenge domain.Challenge, ttl time.Duration) error
	Delete(ctx context.Context, phone string) error
}

type OTPProvider interface {
	Send(ctx context.Context, phone, code string, ttl time.Duration) (*domain.Challenge, error)
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

func generateOTP(length int) (*string, error) {
	limit := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil)
	num, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return nil, fmt.Errorf("generate otp: %w", err)
	}
	code := fmt.Sprintf("%0*d", length, num.Int64())
	return &code, nil
}

const (
	defaultTTL           = 150 * time.Second
	defaultOTPLength int = 6
)

func (s *OTPService) Send(ctx context.Context, phone string) error {
	code, err := generateOTP(defaultOTPLength)
	if err != nil {
		return err
	}
	challenge, err := s.provider.Send(ctx, phone, *code, defaultTTL)
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
