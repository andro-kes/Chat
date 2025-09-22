package services

import "github.com/andro-kes/Chat/chat/internal/repository"

type RoomService interface {
	SendMessage()
}

type roomService struct {
	Repo repository.RoomRepo
}

func NewRoomService() *roomService {
	return &roomService{
		Repo: repository.NewRoomRepo(),
	}
}