package repository

import (
	"context"
	"errors"
	"time"

	"github.com/andro-kes/Chat/auth/internal/database"
	"github.com/andro-kes/Chat/auth/internal/models"
	"github.com/andro-kes/Chat/auth/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type UserRepo struct {
	Pool *pgxpool.Pool
}

func NewUserRepo() *UserRepo {
	return &UserRepo{Pool: database.GetDBPool()}
}

func (dur *UserRepo) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := dur.Pool.QueryRow(
		context.Background(),
		"SELECT id, created_at, updated_at, deleted_at, username, email, password FROM users WHERE email=$1",
		email,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt, &user.Username, &user.Email, &user.Password)

	if err != nil {
		logger.Log.Warn("Пользователь не найден", zap.String("email", email), zap.Error(err))
		return nil, err
	}

	return &user, nil
}

func (dur *UserRepo) CreateUser(user *models.User) error {
	logger.Log.Info("Создаем нового пользователя", zap.String("email", user.Email))

	id := uuid.New()
	_, err := dur.Pool.Exec(
		context.Background(),
		`INSERT INTO users (id, created_at, updated_at, deleted_at, username, email, password) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		id, time.Now(), nil, nil, user.Username, user.Email, user.Password,
	)
	if err != nil {
		logger.Log.Error("Не удалось создать пользователя", zap.Error(err))
		return err
	}
	return nil
}

func (dur *UserRepo) SetPassword(user *models.User) error {
	cmdTag, err := dur.Pool.Exec(
		context.Background(),
		"UPDATE users SET password=$1 WHERE email=$2",
		user.Password, user.Email,
	)
	if err != nil {
		logger.Log.Error("Не удалось обновить пароль", zap.Error(err))
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		logger.Log.Warn("Пароль не обновлён: пользователь не найден", zap.String("email", user.Email))
		return errors.New("не удалось обновить пароль")
	}
	return nil
}