package repository

import (
	"context"

	"github.com/andro-kes/Chat/chat/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepo interface {
	SaveMessage()
	CheckAccess(userId uuid.UUID) error
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

func (rr *roomRepo) CheckAccess(userId uuid.UUID) error {
	var username string
	err := rr.Pool.QueryRow(
		context.Background(),
		"SELECT username FROM users WHERE $1=ANY(users)",
		userId,
	).Scan(&username)

	return err
}