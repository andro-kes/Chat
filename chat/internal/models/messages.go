package models

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        uint      `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"CreatedAt"`
	SenderID  uuid.UUID `db:"sender_id" json:"SenderID"`
	RoomID    uuid.UUID `db:"room_id" json:"RoomID"`
	Content   string    `db:"content" json:"Text"`
}