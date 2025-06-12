package main

import (
	"github.com/andro-kes/Chat/auth/internal/auth"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("../web/templates/*")

	router.GET("/", auth.AuthYandexHandler)
	router.GET("/auth", auth.LoginYandexHandler)

	router.Run(":8000")
}