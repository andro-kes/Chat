package grpc

import (
	context "context"

	"github.com/andro-kes/Chat/auth/internal/services"
	"github.com/andro-kes/Chat/auth/logger"
	"go.uber.org/zap"
)

type authServiceServer struct {
	UnimplementedAuthServiceServer
}

func (ass *authServiceServer) GetUserId(ctx context.Context, t *TokenRequest) (*UserIdResponse, error) {
	tokenService := services.NewTokenService()

	logger.Log.Info(
		"Проверка на авторизацию",
		zap.String("access_token", t.Token),
	)

	claims, err := tokenService.ParseTokenClaims(t.Token)
	if err != nil {
		logger.Log.Error("Не удалось извлечь данные из токена", zap.Error(err))
		return &UserIdResponse{}, err
	}

	userId, err := claims.GetSubject()
	if err != nil {
		logger.Log.Error("Не удалось извлечь данные из токена", zap.Error(err))
		return &UserIdResponse{}, err
	}

	return &UserIdResponse{UserId: userId}, nil
}