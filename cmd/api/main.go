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
	config := config.Load()
	gateway, err := telegramgateway.NewClient(config.Token)
	if err != nil {
		log.Fatalln(err)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})
	otpStore := repository.NewOTPRedis(rdb)
	otpService := service.NewOTPService(gateway, otpStore)
	otpHandler := handler.NewOTPHandler(otpService)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /send-code", otpHandler.Send)
	if err := http.ListenAndServe(config.HTTPAddr, mux); err != nil {
		log.Fatalln(err)
	}
}
