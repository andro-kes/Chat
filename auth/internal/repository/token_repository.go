// ВРЕМЕННО: Пакет repository содержит доступ к БД для refresh токенов.
package repository

import (
	"context"

	"github.com/andro-kes/Chat/auth/internal/database"
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

// Save ВРЕМЕННО: сохраняет refresh token в БД
func (dtr *tokenRepo) Save(userId, tokenId uuid.UUID, tokenString string) error {
	_, err := dtr.Pool.Exec(
		context.Background(),
		"INSERT INTO refresh_tokens (user_id, token_id, token) VALUES ($1, $2, $3)",
		userId, tokenId, tokenString,
	)
	
	if err != nil {
		logger.Log.Error(
			"Не удалось сохранить refresh токен в базу",
			zap.Error(err),
		)
	}

	return err
}

// DeleteByID ВРЕМЕННО: удаляет refresh token по его ID
func (dtr *tokenRepo) DeleteByID(tokenID uuid.UUID) error {
	_, err := dtr.Pool.Exec(
		context.Background(),
		"DELETE FROM refresh_tokens WHERE token_id=$1",
		tokenID,
	)

	if err != nil {
		logger.Log.Error(
			"Не удалось отозвать refresh token",
			zap.Error(err),
		)
	}

	return err
}