package models

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type RefreshTokens struct {
	UserID uint `json:"user_id"`
	TokenID uuid.UUID `json:"token_id"`
	Token string `json:"token"`
}

type RefreshTokensClaims struct {
	UserID uint `json:"user_id"`
	TokenID uuid.UUID `json:"token_id"`
	jwt.StandardClaims
}