package repository

import (
	"context"

	"github.com/andro-kes/Chat/auth/logger"
	"github.com/andro-kes/Chat/auth/internal/models"
	"github.com/andro-kes/Chat/auth/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Логика для работы с базой
type UserRepo interface {
	FindByEmail(email string) *models.User
}

type DBUserRepo struct {
	Pool *pgxpool.Pool
}

func NewUserRepo() *DBUserRepo {
	return &DBUserRepo{
		Pool: database.GetDBPool(),
	}
}

func (dur *DBUserRepo) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := dur.Pool.QueryRow(
		context.Background(), 
		"SELECT * FROM users WHERE email=$1",
		email,
	).Scan(&user)

	if err != nil {
		logger.Log.Warn(
			"Пользователь не найден",
			zap.String("email", user.Email),
			zap.Error(err),
		)
	}

	return &user, err
}