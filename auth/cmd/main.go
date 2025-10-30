package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	authgrpc "github.com/andro-kes/Chat/auth/grpc"
	"github.com/andro-kes/Chat/auth/configs"
	"github.com/andro-kes/Chat/auth/internal/handlers"
	"github.com/andro-kes/Chat/auth/internal/middlewares"
	"github.com/andro-kes/Chat/auth/internal/services"
	"github.com/andro-kes/Chat/auth/logger"
	"go.uber.org/zap"
	grpc "google.golang.org/grpc"
)

func main() {
	logger.Init()
	defer logger.Close()
	configs.InitConfigs()

	// HTTP handlers
	authHandlers := handlers.NewAuthHandlers()
	authMiddlewares := middlewares.NewAuthMiddlewares()

	mux := http.NewServeMux()
	mux.HandleFunc("/yandex_auth", authHandlers.AuthYandexHandler)
	mux.HandleFunc("/auth", authHandlers.LoginYandexHandler)
	mux.HandleFunc("/", authHandlers.LoginPageHandler)
	mux.HandleFunc("/api/login", authHandlers.LoginHandler)
	mux.HandleFunc("/signup_page", authHandlers.SignUPPageHandler)
	mux.HandleFunc("/api/signup", authHandlers.SignUpHandler)
	mux.Handle("/api/logout", authMiddlewares.AuthMiddleware(http.HandlerFunc(authHandlers.LogoutHandler)))

	httpServer := &http.Server{
		Addr:         ":8000",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// gRPC server setup
	grpcAddr := ":50051"
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Log.Fatal("failed to listen for gRPC", zap.Error(err))
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authgrpc.RecoverUnaryInterceptor()),
	)

	authService := authgrpc.NewAuthServiceServer(services.NewTokenService())
	authgrpc.RegisterAuthServiceServer(grpcServer, authService)

	go func() {
		logger.Log.Info("gRPC server starting", zap.String("addr", grpcAddr))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Log.Fatal("gRPC serve error", zap.Error(err))
		}
	}()

	go func() {
		logger.Log.Info("HTTP server starting", zap.String("addr", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("HTTP server error", zap.Error(err))
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	logger.Log.Info("shutdown signal received")

	// HTTP graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Log.Error("HTTP shutdown error", zap.Error(err))
	}

	// gRPC graceful stop
	grpcServer.GracefulStop()

	logger.Log.Info("servers stopped")
}