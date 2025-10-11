package services

import (
	"github.com/andro-kes/Chat/chat/internal/repository"
	"github.com/google/uuid"
)

type RoomService interface {
	SendMessage()
	CheckAccess(userId uuid.UUID) bool
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

func (rs *roomService) CheckAccess(userId uuid.UUID) bool {
	err := rs.Repo.CheckAccess(userId)
	return err == nil
}