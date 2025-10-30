package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/andro-kes/Chat/chat/grpc"
	"github.com/andro-kes/Chat/chat/logger"
	"go.uber.org/zap"
)

// Используем builn-in для совместимости
type UserId string
const UserIDContextKey UserId = "user_id"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			logger.Log.Warn("Отсутствует токен авторизации")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Установим deadline на запрос к auth
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		userId, err := grpc.Client(token)
		if err != nil {
			logger.Log.Warn("Ошибка проверки токена", zap.String("error", err.Error()))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx = context.WithValue(ctx, UserIDContextKey, userId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RecoveryMiddleware реализует middleware для перехвата паник.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
					"Перехвачена паника",
					zap.String("error", msg),
				)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error": "Internal Server Error"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}