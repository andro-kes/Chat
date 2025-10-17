package models

import (
	"time"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/google/uuid"
)

type Room struct {
	ID uuid.UUID `json:"id" db:"id"`
	AdminID uuid.UUID `db:"admin_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
	Name string `json:"name" db:"name"`
	Users pgtype.Array[uuid.UUID] `db:"users"` // хранит только индентификаторы пользователей
}