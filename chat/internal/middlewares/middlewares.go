package middlewares

import (
	"context"
	"fmt"
	"net/http"

	"github.com/andro-kes/Chat/chat/grpc"
	"github.com/andro-kes/Chat/chat/logger"
	"go.uber.org/zap"
)

// Используем plain string ключ для совместимости с getUser
type UserId string
const UserIDKey UserId = "user_id"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Попробовать получить токен из заголовка Authorization
		token := r.Header.Get("Authorization")
		if token != "" {
			// Удалить префикс "Bearer " если есть
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}
		} else {
			// Попробовать из URL-параметра
			token = r.URL.Query().Get("token")
			if token == "" {
				// Попробовать из куки
				cookie, err := r.Cookie("access_token")
				if err == nil {
					token = cookie.Value
				}
			}
		}

		if token == "" {
			logger.Log.Warn("Отсутствует токен авторизации")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userId, err := grpc.Client(token)
		if err != nil {
			logger.Log.Warn("Ошибка проверки токена", zap.String("error", err.Error()))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userId)
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