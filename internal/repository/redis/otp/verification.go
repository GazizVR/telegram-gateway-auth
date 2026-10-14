package redis_otp

import "github.com/gazizvr/telegram-gateway-auth/internal/domain"

type verificationRedisDTO struct {
	CodeHash string `redis:"code_hash"`
	Attempts int    `redis:"attempts"`
}

func (v *verificationRedisDTO) toDomain() *domain.Verification {
	if v == nil {
		return nil
	}
	return &domain.Verification{
		CodeHash: v.CodeHash,
		Attempts: v.Attempts,
	}
}

func toRedisDTO(v domain.Verification) verificationRedisDTO {
	return verificationRedisDTO{
		CodeHash: v.CodeHash,
		Attempts: v.Attempts,
	}
}
