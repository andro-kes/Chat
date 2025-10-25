package services

import (
	"errors"
	"sync"

	"github.com/andro-kes/Chat/chat/internal/models"
	"github.com/andro-kes/Chat/chat/internal/repository"
	"github.com/google/uuid"
)

type ChatService interface {
	AddRoom(room *roomService) error
	DeleteRoom(roomID uuid.UUID) error
	IsActive(roomId uuid.UUID) bool
	GetCurrentRoom(id uuid.UUID) (*models.Room, error)
	CheckAccess(userId uuid.UUID) bool
	CreateRoom(name string, adminID uuid.UUID) error
	GetRoom(roomId uuid.UUID) (RoomService, error)
	GetUserRooms(userId uuid.UUID) ([]models.Room, error)
}

type chatService struct {
	Repo repository.ChatRepo
	ActiveRooms map[uuid.UUID]RoomService
	Mu sync.Mutex 
}

func NewChatService() *chatService {
	return &chatService{
		Repo:        repository.NewChatRepo(),
		ActiveRooms: make(map[uuid.UUID]RoomService),
	}
}

// AddRoom добавляет новую комнату в активные
func (cs *chatService) AddRoom(room *roomService) error {
	if room == nil {
		return errors.New("нельзя добавить nil-комнату")
	}
	cs.Mu.Lock()
	defer cs.Mu.Unlock()
	cs.ActiveRooms[room.ID] = room
	return nil
}

// DeleteRoom удаляет комнату из активных
func (cs *chatService) DeleteRoom(roomID uuid.UUID) error {
	cs.Mu.Lock()
	defer cs.Mu.Unlock()
	if _, exists := cs.ActiveRooms[roomID]; !exists {
		return errors.New("комната не найдена")
	}
	delete(cs.ActiveRooms, roomID)
	return nil
}

// GetRoom возвращает комнату по ID или ошибку
func (cs *chatService) GetRoom(roomId uuid.UUID) (RoomService, error) {
	cs.Mu.Lock()
	defer cs.Mu.Unlock()
	room, ok := cs.ActiveRooms[roomId]
	if !ok {
		return nil, errors.New("комната не найдена")
	}
	return room, nil
}

// GetCurrentRoom возвращает информацию о комнате из репозитория
func (cs *chatService) GetCurrentRoom(id uuid.UUID) (*models.Room, error) {
	return cs.Repo.FindRoomByID(id)
}

// IsActive проверяет, существует ли комната в активных
func (cs *chatService) IsActive(roomId uuid.UUID) bool {
	cs.Mu.Lock()
	defer cs.Mu.Unlock()
	_, ok := cs.ActiveRooms[roomId]
	return ok
}

// CheckAccess проверяет доступ пользователя
func (rs *chatService) CheckAccess(userId uuid.UUID) bool {
	return rs.Repo.CheckAccess(userId) == nil
}

// CreateRoom создает новую комнату и добавляет её в активные
func (rs *chatService) CreateRoom(name string, adminID uuid.UUID) error {
	if err := rs.Repo.CreateRoom(name, adminID); err != nil {
		return err
	}
	newRoom := NewRoomService(uuid.New())
	rs.Mu.Lock()
	rs.ActiveRooms[newRoom.ID] = newRoom
	rs.Mu.Unlock()
	return nil
}

// GetUserRooms возвращает список комнат пользователя
func (cs *chatService) GetUserRooms(userId uuid.UUID) ([]models.Room, error) {
	return cs.Repo.GetUserRooms(userId)
}
