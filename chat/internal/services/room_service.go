package services

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/andro-kes/Chat/chat/internal/models"
	"github.com/andro-kes/Chat/chat/internal/repository"
	"github.com/andro-kes/Chat/chat/logger"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type RoomService interface {
	SendMessage(msg *models.Message) error
	AddUser(userID uuid.UUID, conn *websocket.Conn) error
	RemoveUser(userID uuid.UUID) bool
	GetMessages() ([]models.Message, error)
}

type roomService struct {
	ID        uuid.UUID
	ActiveUsers map[uuid.UUID]*websocket.Conn
	Repo      repository.RoomRepo
	Mu        sync.RWMutex
}

// NewRoomService создает и возвращает новый экземпляр сервиса управления одной комнатой.
func NewRoomService(roomId uuid.UUID) *roomService {
	return &roomService{
		ID:          roomId,
		ActiveUsers: make(map[uuid.UUID]*websocket.Conn),
		Repo:        repository.NewRoomRepo(),
	}
}

// AddUser добавляет пользователя в комнату
func (rs *roomService) AddUser(userID uuid.UUID, conn *websocket.Conn) error {
	rs.Mu.Lock()
	defer rs.Mu.Unlock()

	if _, ok := rs.ActiveUsers[userID]; ok {
		return nil
	}

	rs.ActiveUsers[userID] = conn
	return nil
}

// RemoveUser удаляет пользователя из комнаты
func (rs *roomService) RemoveUser(userID uuid.UUID) bool {
	rs.Mu.Lock()
	defer rs.Mu.Unlock()
	delete(rs.ActiveUsers, userID)

	return len(rs.ActiveUsers) != 0
}

// SendMessage срассылает сообщение всем пользователям
func (rs *roomService) SendMessage(msg *models.Message) error {
	if err := rs.Repo.SaveMessage(msg); err != nil {
		logger.Log.Error("Не удалось сохранить сообщение", zap.Error(err))
		return err
	}

	body, err := json.Marshal(msg)
	if err != nil {
		logger.Log.Error(
			"Не удалось сериализовать сообщение", 
			zap.String("error", err.Error()),
		)
		return err
	}

	rs.Mu.RLock()
	for userID, conn := range rs.ActiveUsers {
		if err := conn.WriteJSON(body); err != nil {
			logger.Log.Warn("Не удалось отправить сообщение пользователю",
				zap.String("user_id", userID.String()),
				zap.Error(err),
			)
			ok := rs.RemoveUser(userID)
			if !ok {
				rs.Mu.RUnlock()
				return errors.New("в комнате не осталось активных пользователей")
			}
		}
	}
	rs.Mu.RUnlock()
	
	return nil
}

// GetMessages возвращает список сообщений комнаты
func (rs *roomService) GetMessages() ([]models.Message, error) {
	return rs.Repo.GetMessages(rs.ID)
}