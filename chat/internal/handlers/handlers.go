package services

import (
	"encoding/json"
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
	RemoveUser(userID uuid.UUID)
	StartWorkers()
	GetMessages() ([]models.Message, error)
}

type roomService struct {
	ID        uuid.UUID
	ActiveUsers map[uuid.UUID]*websocket.Conn
	Repo      repository.RoomRepo
	Mu        sync.RWMutex
	messageChan chan *models.Message
}

// NewRoomService создает и возвращает новый экземпляр сервиса управления одной комнатой.
func NewRoomService(roomId uuid.UUID) *roomService {
	return &roomService{
		ID:          roomId,
		ActiveUsers: make(map[uuid.UUID]*websocket.Conn),
		Repo:        repository.NewRoomRepo(),
		messageChan: make(chan *models.Message, 100),
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
func (rs *roomService) RemoveUser(userID uuid.UUID) {
	rs.Mu.Lock()
	defer rs.Mu.Unlock()
	delete(rs.ActiveUsers, userID)
}

// SendMessage отправляет сообщение всем участникам комнаты
func (rs *roomService) SendMessage(msg *models.Message) error {
	if err := rs.Repo.SaveMessage(msg); err != nil {
		logger.Log.Error("Не удалось сохранить сообщение", zap.Error(err))
		return err
	}

	// Отправляем сообщение через канал
	rs.messageChan <- msg
	return nil
}

// StartWorkers запускает горутины для обработки сообщений
func (rs *roomService) StartWorkers() {
	for range 5 { 
		go func() {
			for {
				msg, ok := <-rs.messageChan
				if !ok {
					return // Канал закрыт
				}

				body, err := json.Marshal(msg)
				if err != nil {
					logger.Log.Error(
						"Не удалось сериализовать сообщение", 
						zap.String("error", err.Error()),
					)
					continue
				}

				rs.Mu.RLock()
				for userID, conn := range rs.ActiveUsers {
					if err := conn.WriteJSON(body); err != nil {
						logger.Log.Warn("Не удалось отправить сообщение пользователю",
							zap.String("user_id", userID.String()),
							zap.Error(err),
						)
						rs.RemoveUser(userID)
					}
				}
				rs.Mu.RUnlock()
			}
		}()
	}
}

// GetMessages возвращает список сообщений комнаты
func (rs *roomService) GetMessages() ([]models.Message, error) {
	return rs.Repo.GetMessages(rs.ID)
}
