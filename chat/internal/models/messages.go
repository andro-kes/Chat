package models

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID uint `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at"`
	SenderID uuid.UUID `db:"sender_id" json:"sender_id"`
	RoomID uuid.UUID `db:"room_id" json:"room_id"`
	Content string `db:"content" json:"content"`
}