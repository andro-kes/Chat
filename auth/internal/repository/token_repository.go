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
	Save(token *models.RefreshTokens) error
}

type DBTokenRepo struct {
	Pool *pgxpool.Pool
}

func NewTokenRepo() *DBTokenRepo {
	return &DBTokenRepo{
		Pool: database.GetDBPool(),
	}
}

func (dtr *DBTokenRepo) Save(userId, tokenId uuid.UUID, tokenString string) error {
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