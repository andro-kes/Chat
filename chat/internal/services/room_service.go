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
	SendMessage(msg *models.Message)
	StartWorkers()
	GetMessages() ([]models.Message, error)
}

type roomService struct {
	ID uuid.UUID

	// ActiveUsers хранит websocket-подключения пользователей в комнате
	ActiveUsers map[uuid.UUID]*websocket.Conn

	Repo repository.RoomRepo

	Mu sync.Mutex
}

// NewRoomService создает и возвращает новый экземпляр сервиса управления одной комнатой.
// Сервис использует репозиторий для взаимодействия с хранилищем данных.
// Пример использования:
//   service := NewRoomService()
func NewRoomService(roomId uuid.UUID) *roomService {
	return &roomService{
		ID: roomId,
		ActiveUsers: make(map[uuid.UUID]*websocket.Conn),
		Repo: repository.NewRoomRepo(),
	}
}

func (rs *roomService) SendMessage(msg *models.Message) {
	body, err := json.Marshal(msg)
	if err != nil {
		logger.Log.Error(
			"Не удалось сериализовать сообщение",
			zap.String("error", err.Error()),
		)
	}

	rs.Mu.Lock()
	for _, conn := range rs.ActiveUsers {
		conn.WriteJSON(body)
	}
	rs.Mu.Unlock()
}

func (rs *roomService) StartWorkers() {
	// Обработка сообщений
}

func (rs *roomService) GetMessages() ([]models.Message, error) {
	return rs.Repo.GetMessages(rs.ID)
}