package services

import (
	"github.com/andro-kes/Chat/chat/internal/models"
	"github.com/andro-kes/Chat/chat/internal/repository"
	"github.com/google/uuid"
)

type RoomService interface {
	SendMessage()
	StartWorkers()
	GetMessages() []models.Message
}

type roomService struct {
	ID uuid.UUID
	ActiveUsers map[uuid.UUID]bool
	Repo repository.RoomRepo
}

// NewRoomService создает и возвращает новый экземпляр сервиса управления одной комнатой.
// Сервис использует репозиторий для взаимодействия с хранилищем данных.
// Пример использования:
//   service := NewRoomService()
func NewRoomService(roomId uuid.UUID) *roomService {
	return &roomService{
		ID: roomId,
		ActiveUsers: make(map[uuid.UUID]bool),
		Repo: repository.NewRoomRepo(),
	}
}

func (*roomService) SendMessage() {
	
}

func (rs *roomService) StartWorkers() {
	// Обработка сообщений
}

func (rs *roomService) GetMessages() []models.Message {
	return rs.Repo.GetMessages(rs.ID)
}