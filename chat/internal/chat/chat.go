package chat

import (
	"net/http"

	"github.com/andro-kes/Chat/shared/models"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Пока что локально доверяю всем серверам
        return true
    },
}

var Manager = RoomManager{
	ActiveRooms: make(map[uint]*models.RoomData),
}











