package main

import (
	"log"
	"net/http"

	"github.com/andro-kes/Chat/chat/internal/handlers"
	"github.com/andro-kes/Chat/chat/internal/middlewares"
)

func main() {
	chatHandlers := handlers.NewChatHandlers()

	// Обёртка хэндлеров в middleware
	http.Handle("/:id/connect", middlewares.RecoveryMiddleware(middlewares.AuthMiddleware(http.HandlerFunc(chatHandlers.ChatHandler))))
	http.Handle("/:id", middlewares.RecoveryMiddleware(middlewares.AuthMiddleware(http.HandlerFunc(chatHandlers.ChatPageHandler))))
	http.Handle("/create", middlewares.RecoveryMiddleware(middlewares.AuthMiddleware(http.HandlerFunc(chatHandlers.CreateRoom))))
	http.Handle("/:id/messages", middlewares.RecoveryMiddleware(middlewares.AuthMiddleware(http.HandlerFunc(chatHandlers.GetRoomMessages))))
	http.Handle("/:id/rooms", middlewares.RecoveryMiddleware(middlewares.AuthMiddleware(http.HandlerFunc(chatHandlers.GetUserRooms))))
	http.Handle("/", middlewares.RecoveryMiddleware(middlewares.AuthMiddleware(http.HandlerFunc(chatHandlers.MainPageHandler))))

	// Запуск сервера
	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}

	// Остановка RabbitMQ
	defer chatHandlers.RabbitManager.Stop()
}
