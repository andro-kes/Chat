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
    UpdateRefreshToken(id uuid.UUID, newToken string) error
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
        logger.Log.Error(
            "Не удалось сохранить refresh токен в базу",
            zap.Error(err),
        )
    }
    return err
}

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

func (dtr *tokenRepo) UpdateRefreshToken(id uuid.UUID, newToken string) error {
    _, err := dtr.Pool.Exec(
        context.Background(),
        "UPDATE tokens SET token=$1 WHERE user_id=$2",
        newToken, id,
    )
    if err != nil {
        logger.Log.Warn(
            "Не удалось обновить рефреш токен",
            zap.String("token",  newToken),
            zap.Any("id", id),
            zap.Error(err),
        )
        return err
    }

    return nil
}

