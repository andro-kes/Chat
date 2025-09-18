// ВРЕМЕННО: Точка входа сервиса аутентификации. Конфигурация роутов и запуск
// HTTP-сервера. Комментарии временные и предназначены для ориентира в период
// рефакторинга.
package main

import (
	"net/http"

	"github.com/andro-kes/Chat/auth/logger"
	"github.com/andro-kes/Chat/auth/internal/handlers"
)

// main ВРЕМЕННО: регистрация HTTP-роутов и запуск сервера
func main() {
	
	authHandlers := handlers.NewAuthHandlers()
	

	http.HandleFunc("/yandex_auth", authHandlers.AuthYandexHandler)
	router.GET("/auth", auth.LoginYandexHandler)

	http.HandleFunc("/", authHandlers.LoginPageHandler)
	http.HandleFunc("/api/login", authHandlers.LoginHandler)
	
	router.GET("/signup_page", auth.SignUPPageHandler)
	router.POST("/api/signup", auth.SignUpHandler)

	router.POST("/logout", auth.LogoutHandler)

	router.PATCH("/api/update", auth.UpdateUser)

	router.Run(":8000")
}