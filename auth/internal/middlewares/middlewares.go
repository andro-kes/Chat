package middlewares

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/andro-kes/Chat/auth/internal/services"
	"github.com/andro-kes/Chat/auth/logger"
	"go.uber.org/zap"
)

type UserId string

const UserIDContextKey UserId = "user_id"

type authMiddlewares struct {
	TokenService services.TokenService
}

func NewAuthMiddlewares() *authMiddlewares {
	return &authMiddlewares{
		TokenService: services.NewTokenService(),
	}
}

func (am *authMiddlewares) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Info("Проверка на авторизованность пользователя")

		// Ожидаем заголовок Authorization: Bearer <token>
		authHeader := r.Header.Get("Authorization")
		var token string
		if authHeader != "" && strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			token = authHeader[7:]
		} else {
			cookie, err := r.Cookie("access_token")
			if err != nil {
				logger.Log.Warn("Access token отсутствует", zap.Error(err))
				http.Error(w, "Access token отсутствует", http.StatusUnauthorized)
				return
			}
			token = cookie.Value
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		userID, err := am.TokenService.ParseAccessToken(token)
		if err != nil {
			logger.Log.Warn("Не удалось извлечь данные из токена", zap.Error(err))
			http.Error(w, "Не удалось извлечь данные из токена", http.StatusUnauthorized)
			return
		}

		ctx = context.WithValue(ctx, UserIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}