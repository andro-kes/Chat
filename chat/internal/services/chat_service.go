package services

import (
	"sync"

	"github.com/andro-kes/Chat/chat/internal/models"
	"github.com/andro-kes/Chat/chat/internal/repository"
	"github.com/google/uuid"
)

type ChatService interface {
	AddRoom(room *roomService)
	DeleteRoom()
	GetCurrentRoom(id uuid.UUID) (*models.Room, error)
}

type chatService struct {
	Repo repository.ChatRepo
	ActiveRooms map[uuid.UUID]RoomService
	Mu *sync.Mutex
}

func NewChatService() *chatService {
	return &chatService{
		Repo: repository.NewChatRepo(),
		ActiveRooms: make(map[uuid.UUID]RoomService),
	}
}

func (cs *chatService) AddRoom(room *roomService) {

	cs.Mu.Lock()
	cs.ActiveRooms[room.ID] = room
}

func (*chatService) DeleteRoom() {
	
}

func (cs *chatService) GetCurrentRoom(id uuid.UUID) (*models.Room, error) {
	return cs.Repo.FindRoomByID(id)
}