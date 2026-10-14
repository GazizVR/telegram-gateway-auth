package redis_otp

import "github.com/gazizvr/telegram-gateway-auth/internal/domain"

type verificationRedisDTO struct {
	OTP      string  `redis:"otp"`
	Attempts int     `redis:"attempts"`
}

func (v verificationRedisDTO) toDomain() domain.Verification {
	return domain.Verification{
		OTP:      v.OTP,
		Attempts: v.Attempts,
	}
}

func toRedisDTO(v domain.Verification) verificationRedisDTO {
	return verificationRedisDTO{
		OTP:      v.OTP,
		Attempts: v.Attempts,
	}
}
