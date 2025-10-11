// ВРЕМЕННО: Точка входа сервиса чата. Настройка роутинга и middlewares.
package main

import (
	"net/http"

	"github.com/andro-kes/Chat/chat/internal/handlers"
)

//
func main() {
	chatHandlers := handlers.NewChatHandlers()

	http.HandleFunc("/", chatHandlers.ChatHandler)
}