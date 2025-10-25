package repository

import (
	"context"

	"github.com/andro-kes/Chat/chat/internal/database"
	"github.com/andro-kes/Chat/chat/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepo interface {
	SaveMessage(msg *models.Message) error
	GetMessages(roomId uuid.UUID) ([]models.Message, error)
}

type roomRepo struct {
	Pool *pgxpool.Pool
}

func NewRoomRepo() *roomRepo {
	return &roomRepo{
		Pool: database.GetDBPool(),
	}
}

// SaveMessage сохраняет сообщение в базе данных
func (rr *roomRepo) SaveMessage(msg *models.Message) error {
	sql := `
		INSERT INTO messages (id, room_id, user_id, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := rr.Pool.Exec(
		context.Background(),
		sql,
		msg.ID,
		msg.RoomID,
		msg.SenderID,
		msg.Content,
		msg.CreatedAt,
	)
	return err
}

// GetMessages возвращает список сообщений комнаты
func (rr *roomRepo) GetMessages(roomId uuid.UUID) ([]models.Message, error) {
	sql := "SELECT id, room_id, user_id, content, created_at FROM messages WHERE room_id = $1 ORDER BY created_at ASC"
	rows, err := rr.Pool.Query(context.Background(), sql, roomId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		if err := rows.Scan(&msg.ID, &msg.RoomID, &msg.SenderID, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}
