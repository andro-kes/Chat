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
	DeleteByUserID(userID uuid.UUID) error
	UpdateRefreshToken(userId uuid.UUID, newToken string) error
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
	_, err := dtr.Pool.Exec(
		context.Background(),
		"INSERT INTO refresh_tokens (user_id, token_id, token) VALUES ($1, $2, $3)",
		userId, tokenId, tokenString,
	)
	if err != nil {
		logger.Log.Error("Не удалось сохранить refresh токен в базу", zap.Error(err))
	}
	return err
}

func (dtr *tokenRepo) DeleteByUserID(userID uuid.UUID) error {
	_, err := dtr.Pool.Exec(
		context.Background(),
		"DELETE FROM refresh_tokens WHERE user_id=$1",
		userID,
	)
	if err != nil {
		logger.Log.Error("Не удалось отозвать refresh token для user", zap.Error(err))
	}
	return err
}

func (dtr *tokenRepo) UpdateRefreshToken(userId uuid.UUID, newToken string) error {
	cmdTag, err := dtr.Pool.Exec(
		context.Background(),
		"UPDATE refresh_tokens SET token=$1 WHERE user_id=$2",
		newToken, userId,
	)
	if err != nil {
		logger.Log.Warn("Не удалось обновить рефреш токен", zap.String("user_id", userId.String()), zap.Error(err))
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		logger.Log.Warn("UpdateRefreshToken не затронуло строк", zap.String("user_id", userId.String()))
	}
	return nil
}