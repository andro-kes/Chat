package repository

import (
	"context"
	"time"

	"github.com/andro-kes/Chat/chat/internal/database"
	"github.com/andro-kes/Chat/chat/internal/models"
	"github.com/andro-kes/Chat/chat/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type ChatRepo interface {
	FindRoomByID(id uuid.UUID) (*models.Room, error)
	CheckAccess(userId uuid.UUID) error
	CreateRoom(name string, adminID uuid.UUID) error
	GetUserRooms(userId uuid.UUID) []models.Room
}

type chatRepo struct {
	Pool *pgxpool.Pool
}

func NewChatRepo() *chatRepo {
	return &chatRepo{
		Pool: database.GetDBPool(),
	}
}

// CreateRoom создает новую комнату и добавляет администратора в список пользователей
func (cr *chatRepo) CreateRoom(name string, adminID uuid.UUID) error {
	roomId := uuid.New()
	now := time.Now()
	users := pgtype.Array[uuid.UUID]{Elements: []uuid.UUID{adminID}}

	sql := `
		INSERT INTO rooms (id, created_at, updated_at, deleted_at, name, users) 
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := cr.Pool.Exec(
		context.Background(),
		sql,
		roomId,
		now,
		now,
		nil,
		name,
		users,
	)
	
	if err != nil {
		logger.Log.Warn(
			"Не удалось создать комнату",
			zap.String("name", name),
			zap.String("admin_id", adminID.String()),
			zap.Error(err),
		)
		return err
	}

	return nil
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

func (rr *chatRepo) CheckAccess(userId uuid.UUID) error {
	var username string
	err := rr.Pool.QueryRow(
		context.Background(),
		"SELECT username FROM users WHERE $1=ANY(users)",
		userId,
	).Scan(&username)

	return err
}

func (rr *chatRepo) GetUserRooms(userId uuid.UUID) []models.Room {
	var rooms []models.Room
	rr.Pool.QueryRow(
		context.Background(),
		"SELECT * FROM rooms WHERE $1=ANY(users)",
		userId,
	).Scan(&rooms)

	return rooms
}