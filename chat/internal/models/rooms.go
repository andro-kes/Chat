package models

import (
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	DeletedAt time.Time `db:"deleted_at"`
	Name string `json:"name" db:"name"`
	Users []uuid.UUID `json:"users" db:"users"`// хранит только индентификаторы пользователей
}