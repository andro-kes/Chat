package models

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID uint
	CreatedAt time.Time
	SenderID uuid.UUID
	RoomID uuid.UUID
	Text string
}