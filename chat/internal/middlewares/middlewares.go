package middlewares

import (
	"context"
	"net/http"

	"github.com/andro-kes/Chat/chat/grpc"
	"github.com/andro-kes/Chat/chat/logger"
)

// Кастомный тип для добавления в контекст
type UserId string

// AuthMiddleware реализует middleware для аутентификации пользователей.
// 
// Функция:
// 1. Извлекает токен из заголовка `Authorization`.
// 2. Вызывает gRPC-клиент для получения идентификатора пользователя.
// 3. Если токен недействителен, отправляет ошибку `404 Not Allowed`.
// 4. Если токен корректен, добавляет `user_id` в контекст запроса.
// 
// Параметры:
//   - next http.Handler: Обработчик, который будет вызван после проверки аутентификации.
// 
// Возвращает:
//   - http.Handler: Новый обработчик с добавленной логикой аутентификации.
// 
// Пример использования:
//   http.Handle("/secure", AuthMiddleware(http.HandlerFunc(YourHandler)))
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		userId, err := grpc.Client(token)
		if err != nil {
			logger.Log.Warn(
				"Не удалось получить токен",
			)
			http.Error(w, "Not allowed", 404)
		}

		var userIdType UserId = "user_id"

		ctx := context.WithValue(r.Context(), userIdType, userId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})	
}