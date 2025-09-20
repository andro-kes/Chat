// ВРЕМЕННО: Точка входа сервиса аутентификации. Конфигурация роутов и запуск
// HTTP-сервера. Комментарии временные и предназначены для ориентира в период
// рефакторинга.
package main

import (
	"net/http"

	"github.com/andro-kes/Chat/auth/configs"
	"github.com/andro-kes/Chat/auth/internal/handlers"
	"github.com/andro-kes/Chat/auth/logger"
	"go.uber.org/zap"
)

// main ВРЕМЕННО: регистрация HTTP-роутов и запуск сервера
func main() {
	logger.Init()
	configs.InitConfigs()
	defer logger.Close()

	authHandlers := handlers.NewAuthHandlers()
	
	http.HandleFunc("/yandex_auth", authHandlers.AuthYandexHandler)
	http.HandleFunc("/auth", authHandlers.LoginYandexHandler)

	http.HandleFunc("/", authHandlers.LoginPageHandler)
	http.HandleFunc("/api/login", authHandlers.LoginHandler)
	
	http.HandleFunc("/signup_page", authHandlers.SignUPPageHandler)
	http.HandleFunc("/api/signup", authHandlers.SignUpHandler)

	http.HandleFunc("/api/logout", authHandlers.LogoutHandler)

	err := http.ListenAndServe("localhost:8000", nil)
	if err != nil {
		logger.Log.Fatal(
			"Не удалось запустить сервер",
			zap.Error(err),
		)
	}
}