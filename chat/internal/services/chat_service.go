package services

import (
	"errors"
	"sync"

	"github.com/andro-kes/Chat/chat/internal/models"
	"github.com/andro-kes/Chat/chat/internal/repository"
	"github.com/google/uuid"
)

type ChatService interface {
	AddRoom(room *roomService)
	DeleteRoom()
	IsActive(roomId uuid.UUID) bool
	GetCurrentRoom(id uuid.UUID) (*models.Room, error)
	CheckAccess(userId uuid.UUID) bool
	CreateRoom(name string, adminID uuid.UUID) error
	GetRoom(roomId uuid.UUID) (RoomService, error)
	GetUserRooms(userId uuid.UUID) []models.Room
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
	cs.Mu.Unlock()
}

func (cs *chatService) GetRoom(roomId uuid.UUID) (RoomService, error) {
	cs.Mu.Lock()
	room, ok := cs.ActiveRooms[roomId]
	cs.Mu.Unlock()
	if !ok {
		return nil, errors.New("комната не найдена")
	}
	return room, nil
}

func (*chatService) DeleteRoom() {
	
}

func (cs *chatService) GetCurrentRoom(id uuid.UUID) (*models.Room, error) {
	return cs.Repo.FindRoomByID(id)
}

func (cs *chatService) IsActive(roomId uuid.UUID) bool {
	_, ok := cs.ActiveRooms[roomId]
	return ok
}

func (rs *chatService) CheckAccess(userId uuid.UUID) bool {
	return rs.Repo.CheckAccess(userId) == nil
}

func (rs *chatService) CreateRoom(name string, adminID uuid.UUID) error {
	if err := rs.Repo.CreateRoom(name, adminID); err != nil {
		return err
	}
	return nil
}

func (cs *chatService) GetUserRooms(userId uuid.UUID) []models.Room {
	return cs.Repo.GetUserRooms(userId)
}