package main

import (
	"log"
	
	"github.com/andro-kes/Chat/auth/internal/auth"
	"github.com/andro-kes/Chat/shared/middlewares"

	"github.com/gin-gonic/gin"
)

func main() {
	defer func() {
		sqlDB, err := middlewares.DB.DB()
		if err != nil {
			log.Fatalln("Ошибка при попытке закрыть базу данных: %м", err)
		}
		sqlDB.Close()
		log.Println("Соединение с базой данных разорвано")
	} ()

	router := gin.Default()
	router.Use(middlewares.DBMiddleWare())
	router.LoadHTMLGlob("/app/web/templates/*")

	router.GET("/yandex_auth", auth.AuthYandexHandler)
	router.GET("/auth", auth.LoginYandexHandler)

	router.GET("/", auth.LoginPageHandler)
	router.POST("/api/login", auth.LoginHandler)
	
	router.GET("/signup_page", auth.SignUPPageHandler)
	router.POST("/api/signup", auth.SignUpHandler)

	router.POST("/logout", auth.LogoutHandler)

	router.PATCH("/api/update", auth.UpdateUser)

	router.Run(":8000")
}