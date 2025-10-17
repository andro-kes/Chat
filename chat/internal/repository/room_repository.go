package repository

import (
	"context"

	"github.com/andro-kes/Chat/chat/internal/database"
	"github.com/andro-kes/Chat/chat/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepo interface {
	SaveMessage()
	GetMessages(roomId uuid.UUID) []models.Message
}

type roomRepo struct {
	Pool *pgxpool.Pool
}

func NewRoomRepo() *roomRepo {
	return &roomRepo{
		Pool: database.GetDBPool(),
	}
}

func (rr *roomRepo) SaveMessage() {
	
}

func (rr *roomRepo) GetMessages(roomId uuid.UUID) []models.Message {
	var messages []models.Message
	rr.Pool.QueryRow(
		context.Background(),
		"SELECT * FROM messages WHERE room_id=$1",
		roomId,
	).Scan(&messages)

	return messages
}