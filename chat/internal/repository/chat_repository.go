package repository

import (
	"context"

	"github.com/andro-kes/Chat/chat/internal/database"
	"github.com/andro-kes/Chat/chat/internal/models"
	"github.com/andro-kes/Chat/chat/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type ChatRepo interface {
	CreateRoom()
	FindRoomByID(id uuid.UUID) (*models.Room, error)
}

type chatRepo struct {
	Pool *pgxpool.Pool
}

func NewChatRepo() *chatRepo {
	return &chatRepo{
		Pool: database.GetDBPool(),
	}
}

func (*chatRepo) CreateRoom() {
	
}

func (cr *chatRepo) FindRoomByID(id uuid.UUID) (*models.Room, error) {
	sql := `SELECT name, users FROM rooms WHERE id = $1`
	var room models.Room
	err := cr.Pool.QueryRow(
		context.Background(),
		sql,
		id,
	).Scan(&room.Name, &room.Users)
	if err != nil {
		logger.Log.Warn(
			"Не удалось найти комнату",
			zap.String("id", id.String()),
			zap.Error(err),
		)
		return &models.Room{}, err
	}

	return &room, nil
}