package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID uuid.UUID
	CreatedAt time.Time
	DeletedAt time.Time
	UpdatedAt time.Time
	Username string `json:"username"`
	Email string `json:"email"`
	Password string `json:"password"`
}