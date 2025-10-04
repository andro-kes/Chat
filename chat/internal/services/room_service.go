package services

import (
	"github.com/andro-kes/Chat/chat/internal/repository"
	"github.com/google/uuid"
)

type RoomService interface {
	SendMessage()
}

type roomService struct {
	ID uuid.UUID
	Repo repository.RoomRepo
}

func NewRoomService() *roomService {
	return &roomService{
		Repo: repository.NewRoomRepo(),
	}
}

func (*roomService) SendMessage() {
	
}