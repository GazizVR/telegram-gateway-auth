// Command server runs the OTP verification HTTP service.
//
// It sends codes through Telegram Gateway, falls back to Eskiz SMS,
// and stores code hashes in Redis or Valkey. Configuration is read from
// environment variables; see package config.
package main

import (
	"log"
	"log/slog"
	"net"
	"net/http"

	telegramgateway "github.com/gazizvr/go-telegram-gateway"
	"github.com/gazizvr/telegram-gateway-auth/internal/adapter"
	"github.com/gazizvr/telegram-gateway-auth/internal/client/eskiz"
	"github.com/gazizvr/telegram-gateway-auth/internal/config"
	"github.com/gazizvr/telegram-gateway-auth/internal/handler"
	redis_otp "github.com/gazizvr/telegram-gateway-auth/internal/repository/redis/otp"
	"github.com/gazizvr/telegram-gateway-auth/internal/server"
	"github.com/gazizvr/telegram-gateway-auth/internal/service"
	"github.com/gazizvr/telegram-gateway-auth/internal/web"
	"github.com/gazizvr/telegram-gateway-auth/otpservice"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
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
	var smsSender service.OTPSender
	if config.EskizEmail == "" || config.EskizPassword == "" {
		slog.Warn("ESKIZ_EMAIL or ESKIZ_PASSWORD is not set, SMS fallback is disabled")
		smsSender = adapter.NewSMSSender()
	} else {
		smsSender = eskiz.New(config.EskizEmail, config.EskizPassword)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
	})
	redisStore := redis_otp.NewStore(rdb)
	otpService := service.NewOTPService(telegramSender, smsSender, redisStore, redisStore, redisStore, []byte(config.SecretKey))
	go func() {
		service := server.NewOTPServiceServer(otpService)
		ln, err := net.Listen("tcp", ":1234")
		if err != nil {
			log.Fatal(err)
		}
		server := grpc.NewServer()
		otpservice.RegisterOTPServiceServer(server, service)
		if err := server.Serve(ln); err != nil {
			log.Fatal(err)
		}
	}()
	otpHandler := handler.NewOTPHandler(otpService)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/send-code", otpHandler.Send)
	mux.HandleFunc("POST /v1/verify-code", otpHandler.Verify)
	if config.WebDemo {
		webHandler := web.New()
		webHandler.Routes(mux)
	}
	if err := http.ListenAndServe(config.HTTPAddr, mux); err != nil {
		log.Fatalln(err)
	}
}
