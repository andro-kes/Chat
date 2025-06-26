package chat

import (
	"sync"
	"github.com/andro-kes/Chat/shared/models"
)

type RoomManager struct {
	ActiveRooms map[uint]*models.RoomData
	Mu sync.RWMutex
}

func (roomManager *RoomManager) AddRoom(roomData *models.RoomData) {
	roomManager.Mu.Lock()
	defer roomManager.Mu.Unlock()
	roomManager.ActiveRooms[roomData.Room.ID] = roomData
}

func (roomManager *RoomManager) Delete(roomData *models.RoomData) {
	roomManager.Mu.Lock()
	defer roomManager.Mu.Unlock()
	delete(roomManager.ActiveRooms, roomData.Room.ID)
}

