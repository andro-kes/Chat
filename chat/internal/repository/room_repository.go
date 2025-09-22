package repository

import (
	"github.com/andro-kes/Chat/chat/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepo interface {
	SaveMessage()
}

type roomRepo struct {
	Pool *pgxpool.Pool
}

func NewRoomRepo() *roomRepo {
	return &roomRepo{
		Pool: database.GetDBPool(),
	}
}

func (r *roomRepo) SaveMessage() {
	
}