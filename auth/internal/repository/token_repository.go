package repository

import (
	"context"

	"github.com/andro-kes/Chat/auth/internal/database"
	"github.com/andro-kes/Chat/auth/internal/models"
	"github.com/andro-kes/Chat/auth/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type TokenRepo interface {
	Save(userId, tokenId uuid.UUID, tokenString string) error
	DeleteByID(tokenID uuid.UUID) error
}

type tokenRepo struct {
	Pool *pgxpool.Pool
}

func NewTokenRepo() *tokenRepo {
	return &tokenRepo{
		Pool: database.GetDBPool(),
	}
}

func (dtr *tokenRepo) Save(userId, tokenId uuid.UUID, tokenString string) error {
	var token models.RefreshTokens
	err := dtr.Pool.QueryRow(
		context.Background(),
		"INSERT INTO refresh_tokens (user_id, token_id, token) VALUES ($1, $2, $3)",
		userId, tokenId, tokenString,
	).Scan(&token)
	
	if err != nil {
		logger.Log.Error(
			"Не удалось сохранить refresh токен в базу",
			zap.Error(err),
		)
	}

	return err
}

func (dtr *tokenRepo) DeleteByID(tokenID uuid.UUID) error {
	var token models.RefreshTokens
	err := dtr.Pool.QueryRow(
		context.Background(),
		"DELETE FROM refresh_tokens WHERE token_id=$1",
		tokenID,
	).Scan(&token)

	if err != nil {
		logger.Log.Error(
			"Не удалось отозвать refresh token",
			zap.Error(err),
		)
	}

	return err
}