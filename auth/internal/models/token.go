package models

import (
	"time"

	"github.com/google/uuid"
)

type RefreshTokens struct {
	UserID uuid.UUID `json:"user_id" db:"user_id"`
	TokenID uuid.UUID `json:"token_id" db:"token_id"`
	Token string `json:"token" db:"token"`
	CreatedAt time.Time `db:"created_at"`
}