package grpc

import (
	context "context"

	"github.com/andro-kes/Chat/auth/logger"
	"go.uber.org/zap"
)

type authServiceServer struct {
	UnimplementedAuthServiceServer
}

func (ass *authServiceServer) GetUserId(ctx context.Context, t *TokenRequest) (*UserIdResponse, error) {
	logger.Log.Info(
		"Проверка на авторизацию",
		zap.String("access_token", t.Token),
	)

	
}