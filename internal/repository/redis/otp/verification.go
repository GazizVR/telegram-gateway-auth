package redis_otp

import "github.com/gazizvr/telegram-gateway-auth/internal/domain"

type verificationRedisDTO struct {
	ID       *string `redis:"id"`
	OTP      string  `redis:"otp"`
	Attempts int     `redis:"attempts"`
}

func (v verificationRedisDTO) toDomain() domain.Verification {
	id := domain.VerificationID(*v.ID)
	return domain.Verification{
		ID:       &id,
		OTP:      v.OTP,
		Attempts: v.Attempts,
	}
}

func toRedisDTO(v domain.Verification) verificationRedisDTO {
	id := string(*v.ID)
	return verificationRedisDTO{
		ID:       &id,
		OTP:      v.OTP,
		Attempts: v.Attempts,
	}
}
