package main

import (
	"log"
	"net/http"

	telegramgateway "github.com/gazizvr/go-telegram-gateway"
	"github.com/gazizvr/telegram-gateway-auth/internal/config"
	"github.com/gazizvr/telegram-gateway-auth/internal/handler"
	"github.com/gazizvr/telegram-gateway-auth/internal/repository"
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
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
	})
	otpStore := repository.NewOTPRedis(rdb)
	otpService := service.NewOTPService(gateway, otpStore)
	otpHandler := handler.NewOTPHandler(otpService)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/send-code", otpHandler.Send)
	mux.HandleFunc("POST /v1/verify-code", otpHandler.Verify)
	if err := http.ListenAndServe(config.HTTPAddr, mux); err != nil {
		log.Fatalln(err)
	}
}
