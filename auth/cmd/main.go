// ВРЕМЕННО: Точка входа сервиса аутентификации. Конфигурация роутов и запуск
// HTTP-сервера. Комментарии временные и предназначены для ориентира в период
// рефакторинга.
package main

import (
	"net/http"

	"github.com/andro-kes/Chat/auth/configs"
	"github.com/andro-kes/Chat/auth/internal/handlers"
	"github.com/andro-kes/Chat/auth/internal/middlewares"
	"github.com/andro-kes/Chat/auth/logger"
	"go.uber.org/zap"
)

// main ВРЕМЕННО: регистрация HTTP-роутов и запуск сервера
func main() {
	logger.Init()
	configs.InitConfigs()
	defer logger.Close()

	authHandlers := handlers.NewAuthHandlers()
	authMiddlewares := middlewares.NewAuthMiddlewares()

	mux := http.NewServeMux()
	
	mux.HandleFunc("/yandex_auth", authHandlers.AuthYandexHandler)
	mux.HandleFunc("/auth", authHandlers.LoginYandexHandler)

	mux.HandleFunc("/", authHandlers.LoginPageHandler)
	mux.HandleFunc("/api/login", authHandlers.LoginHandler)
	
	mux.HandleFunc("/signup_page", authHandlers.SignUPPageHandler)
	mux.HandleFunc("/api/signup", authHandlers.SignUpHandler)

	logoutHandler := authMiddlewares.AuthMiddleware(http.HandlerFunc(authHandlers.LogoutHandler))
	mux.Handle("/api/logout", logoutHandler)

	err := http.ListenAndServe("localhost:8000", mux)
	if err != nil {
		logger.Log.Fatal(
			"Не удалось запустить сервер",
			zap.Error(err),
		)
	}
}