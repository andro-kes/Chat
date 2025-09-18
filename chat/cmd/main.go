// ВРЕМЕННО: Точка входа сервиса чата. Настройка роутинга и middlewares.
package main

import (
	"github.com/andro-kes/Chat/chat/internal/chat"
	"github.com/andro-kes/Chat/shared/middlewares"

	"github.com/gin-gonic/gin"
)

// main ВРЕМЕННО: инициализирует gin, подключает middlewares и регистрирует роуты
func main() {
	router := gin.Default()
	router.Use(middlewares.DBMiddleWare())
	router.Use(middlewares.IsAuthMiddleware())
	router.LoadHTMLGlob("/app/web/templates/*")

	router.GET("/", chat.MainPageHandler)
	router.GET("/api/rooms", chat.GetUserRooms)
	router.GET("/:id", chat.ChatPageHandler)
	router.GET("/api/room/:id/messages", chat.GetRoomMessages)
	router.GET("/:id/ws", chat.ChatHandler)
	router.POST("/create_room", chat.CreateRoom)
	router.POST("/api/:id/add_user", chat.AddUserToRoom)

	router.Run(":8080")
}