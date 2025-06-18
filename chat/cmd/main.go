package main

import (
	"github.com/andro-kes/Chat/chat/internal/chat"
	"github.com/andro-kes/Chat/shared/middlewares"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.Use(middlewares.DBMiddleWare())
	router.Use(middlewares.IsAuthMiddleware())
	router.LoadHTMLGlob("/app/web/templates/*")

	router.GET("/", chat.MainPageHandler)
	router.GET("/:id", chat.ChatPageHandler)
	router.GET("/:id/ws")

	router.Run(":8080")
}