package models

import (
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
	Name string
}