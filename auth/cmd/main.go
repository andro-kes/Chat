package main

import (
	"net/http"

	"github.com/andro-kes/Chat/auth/logger"
	"github.com/andro-kes/Chat/auth/internal/handlers"
)

func main() {
	logger.Init()

	defer logger.Close()

	

	http.HandleFunc("/yandex_auth", handlers.AuthYandexHandler)
	router.GET("/auth", auth.LoginYandexHandler)

	router.GET("/", auth.LoginPageHandler)
	http.HandleFunc("/api/login", auth.LoginHandler)
	
	router.GET("/signup_page", auth.SignUPPageHandler)
	router.POST("/api/signup", auth.SignUpHandler)

	router.POST("/logout", auth.LogoutHandler)

	router.PATCH("/api/update", auth.UpdateUser)

	router.Run(":8000")
}