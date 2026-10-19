package server

import (
	"context"

	otp_service "github.com/gazizvr/telegram-gateway-auth/otpservice"
)

type OTPService interface {
	Send(ctx context.Context, phone string) error
	Verify(ctx context.Context, phone, code string) error
}

type OTPServiceServer struct {
	otp_service.UnimplementedOTPServiceServer
	src OTPService
}

func NewOTPServiceServer(src OTPService) *OTPServiceServer {
	return &OTPServiceServer{
		src: src,
	}
}

func (s *OTPServiceServer) Send(
	ctx context.Context,
	req *otp_service.SendRequest,
) (*otp_service.SendResponse, error) {
	if err := s.src.Send(ctx, req.Phone); err != nil {
		return nil, err
	}
	return &otp_service.SendResponse{}, nil
}

func (s *OTPServiceServer) Verify(
	ctx context.Context,
	req *otp_service.VerifyRequest,
) (*otp_service.VerifyResponse, error) {
	if err := s.src.Verify(ctx, req.Phone, req.Code); err != nil {
		return nil, err
	}
	return &otp_service.VerifyResponse{}, nil
}
