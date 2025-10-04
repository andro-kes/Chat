// ВРЕМЕННО: Точка входа сервиса чата. Настройка роутинга и middlewares.
package main

import (
	"net/http"

	"github.com/andro-kes/Chat/chat/internal/handlers"
	"github.com/andro-kes/Chat/chat/internal/services"
)

//
func main() {
	chatHandlers := handlers.NewChatHandlers()

	http.HandleFunc("/", chatHandlers.)
}