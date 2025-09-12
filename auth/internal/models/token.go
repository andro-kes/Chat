package models

import (
	"github.com/google/uuid"
)

type RefreshTokens struct {
	UserID uuid.UUID `json:"user_id"`
	TokenID uuid.UUID `json:"token_id"`
	Token string `json:"token"`
}