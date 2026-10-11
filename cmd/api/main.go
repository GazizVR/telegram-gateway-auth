package main

import (
	"log"
	"net/http"

	telegramgateway "github.com/gazizvr/go-telegram-gateway"
	"github.com/gazizvr/telegram-gateway-auth/internal/config"
	"github.com/gazizvr/telegram-gateway-auth/internal/handler"
	"github.com/gazizvr/telegram-gateway-auth/internal/service"
)

func main() {
	config := config.Load()
	gateway, err := telegramgateway.NewClient(config.Token)
	if err != nil {
		log.Fatalln(err)
	}
	otpService := service.NewOTPService(*gateway)
	otpHandler := handler.NewOTPHandler(otpService)
	mux := http.NewServeMux()
	mux.HandleFunc("/send-code", otpHandler.Send)
	if err := http.ListenAndServe("0.0.0.0:7777", mux); err != nil {
		log.Fatalln(err)
	}
}
