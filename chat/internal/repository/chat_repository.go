package repository

import (
	"context"
	"errors"
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
	GetUserRooms(userId uuid.UUID) ([]models.Room, error)
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
	users := pgtype.Array[uuid.UUID]{
		Elements: []uuid.UUID{adminID},
		Valid:    true,
	}

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

// FindRoomByID возвращает комнату по ID или ошибку
func (cr *chatRepo) FindRoomByID(id uuid.UUID) (*models.Room, error) {
	sql := `SELECT id, name, users FROM rooms WHERE id = $1 AND deleted_at IS NULL`
	var room models.Room
	err := cr.Pool.QueryRow(
		context.Background(),
		sql,
		id,
	).Scan(&room.ID, &room.Name, &room.Users)
	if err != nil {
		logger.Log.Warn(
			"Не удалось найти комнату",
			zap.String("id", id.String()),
			zap.Error(err),
		)
		return nil, err
	}

	return &room, nil
}

// CheckAccess проверяет, имеет ли пользователь доступ к комнате
func (rr *chatRepo) CheckAccess(userId uuid.UUID) error {
	var username string
	err := rr.Pool.QueryRow(
		context.Background(),
		"SELECT username FROM users u JOIN room_users ru ON u.id = ru.user_id WHERE ru.room_id = ANY(SELECT id FROM rooms WHERE deleted_at IS NULL) AND ru.user_id = $1",
		userId,
	).Scan(&username)

	if err != nil {
		return errors.New("не удалось выполнить запрос")
	}

	if username == "" {
		return errors.New("нет доступа")
	}

	return nil
}

// GetUserRooms возвращает список комнат, к которым имеет доступ пользователь
func (rr *chatRepo) GetUserRooms(userId uuid.UUID) ([]models.Room, error) {
	sql := `
		SELECT r.id, r.name, r.users 
		FROM rooms r
		JOIN room_users ru ON r.id = ru.room_id
		WHERE ru.user_id = $1 AND r.deleted_at IS NULL
	`

	rows, err := rr.Pool.Query(context.Background(), sql, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []models.Room
	for rows.Next() {
		var room models.Room
		if err := rows.Scan(&room.ID, &room.Name, &room.Users); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rooms, nil
}
