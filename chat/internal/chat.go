package chat

import (
	"github.com/andro-kes/Chat/shared/models"
)

var Manager = RoomManager{
	ActiveRooms: make(map[uint]*models.RoomData),
}











