package grpc

import (
	"context"
	"fmt"

	"github.com/andro-kes/Chat/auth/internal/services"
	"github.com/andro-kes/Chat/auth/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// authServiceServer реализует интерфейс AuthServiceServer.
// Важно: встраиваем (embed) UnimplementedAuthServiceServer для forward-совместимости.
type authServiceServer struct {
	UnimplementedAuthServiceServer
	tokenService services.TokenService
}

// NewAuthServiceServer создает новый экземпляр сервера gRPC.
func NewAuthServiceServer(tokenService services.TokenService) *authServiceServer {
	return &authServiceServer{
		tokenService: tokenService,
	}
}

// GetUserId возвращает идентификатор пользователя из токена.
func (ass *authServiceServer) GetUserId(ctx context.Context, t *TokenRequest) (*UserIdResponse, error) {
	// Не логируем сам токен (чувствительные данные)
	logger.Log.Info("Проверка авторизации (gRPC)")

	claims, err := ass.tokenService.ParseTokenClaims(t.Token)
	if err != nil {
		logger.Log.Error("Не удалось извлечь данные из токена", zap.Error(err))
		return nil, status.Errorf(codes.Unauthenticated, "Ошибка проверки токена")
	}

	// RegisteredClaims содержит Subject как строку
	userId := claims.Subject
	if userId == "" {
		logger.Log.Error("Токен не содержит subject")
		return nil, status.Errorf(codes.Unauthenticated, "Некорректный токен")
	}

	return &UserIdResponse{UserId: userId}, nil
}

// RecoverUnaryInterceptor - unary interceptor для перехвата паник в gRPC.
func RecoverUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				var msg string
				switch v := r.(type) {
				case error:
					msg = v.Error()
				default:
					msg = fmt.Sprintf("%v", v)
				}

				logger.Log.Error(
					"Перехвачена паника в gRPC-обработчике",
					zap.String("method", info.FullMethod),
					zap.String("error", msg),
				)

				resp = nil
				err = status.Errorf(codes.Internal, "Внутренняя ошибка сервера")
			}
		}()
		return handler(ctx, req)
	}
}