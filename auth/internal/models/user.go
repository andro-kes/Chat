// ВРЕМЕННО: Пакет models содержит структуры домена auth. Комментарии временные.
package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID uuid.UUID `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	DeletedAt *time.Time `db:"deleted_at"`
	UpdatedAt *time.Time `db:"updated_at"`
	Username string `json:"username" db:"username"`
	Email string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
}