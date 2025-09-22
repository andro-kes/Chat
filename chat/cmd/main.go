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
	router.LoadHTMLGlob("/app/web/templates/*")

    // Публичный endpoint для обмена токенов на серверные HttpOnly куки
    router.GET("/auth/exchange", chat.AuthExchangeHandler)

    // Защищенные маршруты
    auth := router.Group("/")
    auth.Use(middlewares.IsAuthMiddleware())
    {
        auth.GET("/", chat.MainPageHandler)
        auth.GET("/api/rooms", chat.GetUserRooms)
        auth.GET("/:id", chat.ChatPageHandler)
        auth.GET("/api/room/:id/messages", chat.GetRoomMessages)
        auth.GET("/:id/ws", chat.ChatHandler)
        auth.POST("/create_room", chat.CreateRoom)
        auth.POST("/api/:id/add_user", chat.AddUserToRoom)
    }

	router.Run(":8080")
}