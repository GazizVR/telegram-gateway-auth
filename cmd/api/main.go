package main

import (
	"log"
	"net/http"

	telegramgateway "github.com/gazizvr/go-telegram-gateway"
	"github.com/gazizvr/telegram-gateway-auth/internal/adapter"
	"github.com/gazizvr/telegram-gateway-auth/internal/config"
	"github.com/gazizvr/telegram-gateway-auth/internal/handler"
	redis_otp "github.com/gazizvr/telegram-gateway-auth/internal/repository/redis/otp"
	"github.com/gazizvr/telegram-gateway-auth/internal/service"
	"github.com/redis/go-redis/v9"
)

func main() {
	config, err := config.Load()
	if err != nil {
		log.Fatalln(err)
	}
	gateway, err := telegramgateway.NewClient(config.Token)
	if err != nil {
		log.Fatalln(err)
	}
	telegramSender := adapter.NewTelegramSender(gateway)
	smsSender := adapter.NewSMSSender()
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
	})
	otpStore := redis_otp.NewStore(rdb)
	otpService := service.NewOTPService(telegramSender, smsSender, otpStore, []byte(config.SecretKey))
	otpHandler := handler.NewOTPHandler(otpService)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/send-code", otpHandler.Send)
	mux.HandleFunc("POST /v1/verify-code", otpHandler.Verify)
	if err := http.ListenAndServe(config.HTTPAddr, mux); err != nil {
		log.Fatalln(err)
	}
}
