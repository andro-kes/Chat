package middlewares

import (
	"context"
	"net/http"
	"time"

	"github.com/andro-kes/Chat/auth/internal/services"
	"github.com/andro-kes/Chat/auth/logger"
	"go.uber.org/zap"
)

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

		token := r.Header.Get("Authentication") 

		cookie, err := r.Cookie("access_token")
		if err != nil {
			logger.Log.Warn(
				"Отсутствует access_token",
				zap.Error(err),
			)
			http.Error(w, "Access token отсутствует", http.StatusUnauthorized)
			return
		}
		accessToken := cookie.Value

		if accessToken != token {
			logger.Log.Warn(
				"access токены не совпали",
			)
			http.Error(w, "Требуется авторизация", http.StatusUnauthorized)
			return
		}

		claims, err := am.TokenService.ParseTokenClaims(accessToken)
		if err != nil {
			logger.Log.Error("Не удалось извлечь данные из токена", zap.Error(err))
			http.Error(w, "Не удалось извлечь данные из токена", http.StatusBadRequest)
			return
		}

		if time.Now().After(claims.ExpiresAt.Time) {
			logger.Log.Info("Обновление токенов")
			refreshTokenCookie, err := r.Cookie("refresh_token")
			if err != nil {
				logger.Log.Error("Refresh token отсутствует", zap.Error(err))
				http.Error(w, "Refresh token отсутствует", http.StatusUnauthorized)
				return
			}

			refreshTokenClaims, err := am.TokenService.ParseTokenClaims(accessToken)
			if err != nil {
				logger.Log.Error("Не удалось извлечь данные из токена", zap.Error(err))
				http.Error(w, "Не удалось извлечь данные из токена", http.StatusBadRequest)
				return
			}
			if time.Now().After(refreshTokenClaims.ExpiresAt.Time) {
				logger.Log.Error("Период действия Refresh token истек")
				http.Error(w, "Период действия Refresh token истек", http.StatusUnauthorized)
				return
			}

			newToken, err := am.TokenService.UpdateRefreshToken(refreshTokenCookie.Value)
			if err != nil {
				logger.Log.Error("Не удалось обновить refresh token", zap.Error(err))
				http.Error(w, "Не удалось обновить refresh token", http.StatusUnauthorized)
				return
			}
			cookie := &http.Cookie{
				Name:     "refresh_token",
				Value:    newToken,
				Expires:  time.Now().Add(720 * time.Hour),
				Path:     "/",
				HttpOnly: true, // Доступ только через HTTP, защита от XSS
				Secure:   true, // Только HTTPS
				SameSite: http.SameSiteStrictMode, // Защита от CSRF
			}
			http.SetCookie(w, cookie)

			newToken, err = am.TokenService.UpdateAccessToken(accessToken)
			if err != nil {
				http.Error(w, "Не удалось обновить access token", http.StatusUnauthorized)
				return
			}
			cookie = &http.Cookie{
				Name:     "access_token",
				Value:    newToken,
				Expires:  time.Now().Add(5 * time.Minute),
				Path:     "/",
				HttpOnly: true, // Доступ только через HTTP, защита от XSS
				Secure:   true, // Только HTTPS
				SameSite: http.SameSiteStrictMode, // Защита от CSRF
			}
			http.SetCookie(w, cookie)
		}

		userID, err := am.TokenService.ParseAccessToken(token)
		if err != nil {
			http.Error(w, "Не удалось извлечь данные из токена", http.StatusBadRequest)
			return
		}

		logger.Log.Info("Добавление пользователя в контекст")
		ctx := context.WithValue(r.Context(), "user_id", userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}