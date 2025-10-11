package middlewares

import (
	"context"
	"net/http"

	"github.com/andro-kes/Chat/chat/grpc"
	"github.com/andro-kes/Chat/chat/logger"
)

type UserId string

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