// ВРЕМЕННО: Пакет repository инкапсулирует доступ к БД (users).
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

// UserRepo ВРЕМЕННО: интерфейс доступа к данным пользователя
type UserRepo interface {
	FindByEmail(email string) (*models.User, error)
	CreateUser(user *models.User) error
	SetPassword(user *models.User) error
}

type DBUserRepo struct {
	Pool *pgxpool.Pool
}

func NewUserRepo() UserRepo {
	return &DBUserRepo{
		Pool: database.GetDBPool(),
	}
}

// FindByEmail ВРЕМЕННО: ищет пользователя по email
func (dur *DBUserRepo) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := dur.Pool.QueryRow(
		context.Background(), 
		"SELECT id, created_at, updated_at, deleted_at, username, email, password FROM users WHERE email=$1",
		email,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt, &user.Username, &user.Email, &user.Password)

	if err != nil {
		logger.Log.Warn(
			"Пользователь не найден",
			zap.String("email", email),
			zap.Error(err),
		)
	}

	return &user, err
}

// CreateUser ВРЕМЕННО: создает нового пользователя
func (dur *DBUserRepo) CreateUser(user *models.User) error {
	logger.Log.Info(
		"Создаем нового пользователя",
		zap.String("email", user.Email),
	)
	
	query := `
	INSERT INTO users 
	(id, created_at, updated_at, deleted_at, username, email, password)
	VALUES
	($1, $2, $3, $4, $5, $6, $7)
	`

	id, err := uuid.NewUUID()
	if err != nil {
		logger.Log.Error(
			"Не удалось создать uuid",
			zap.Error(err),
		)
		return err
	}

	var newUser models.User
	err = dur.Pool.QueryRow(
		context.Background(),
		query,
		id, time.Now(), nil, nil, user.Username, user.Email, user.Password,
	).Scan(&newUser)

	if err != nil {
		logger.Log.Error(
			"Не удалось создать пользователя",
			zap.Error(err),
		)
	}

	return err
}


// SetPassword ВРЕМЕННО: обновляет пароль пользователя
func (dur *DBUserRepo) SetPassword(user *models.User) error {
	row := dur.Pool.QueryRow(
		context.Background(),
		"UPDATE users SET password=$1 WHERE email=$2",
		user.Password, user.Email,
	)
	if row != nil {
		logger.Log.Warn(
			"Не удалось добавить пароль полльзователю",
			zap.String("password", user.Password),
		)
		return errors.New("Не удалось обновить пароль")
	}
	return nil
}