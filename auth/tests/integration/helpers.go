package auth_tests

import (
	"testing"

	"github.com/andro-kes/Chat/auth/internal/database"
	"github.com/andro-kes/Chat/auth/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func CreatePool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	
	config, err := pgxpool.ParseConfig("postgresql:///memory:?mode=memory&cache=shared")
	if err != nil {
		logger.Log.Fatal(
			"Не удалось подключить к тсетовой бд",
			zap.Error(err),
		)
	}

	pool, err := pgxpool.NewWithConfig(t.Context(), config)
	if err != nil {
		logger.Log.Fatal(
			"Не удалось создать пул",
			zap.Error(err),
		)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}

func makeMigrations(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	fixtures := []string{
		`CREATE TABLE users(
			id UUID PRIMARY KEY,
			username VARCHAR(50),
			email VARCHAR(50),
			password VARCHAR(50),
		)
		`,
		`CREATE TABLE tokens(
			token_id UUID PRIMARY KEY,
			user_id UUID,
			token VARCHAR(100),
		)
		`,
	}

	for _, fixture := range fixtures {
		_, err := pool.Exec(t.Context(), fixture)
		if err != nil {
			logger.Log.Fatal(
				"Не удалось провести миграцию",
				zap.Error(err),
			)
		}
	}
}

func createTestUser(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	sql := "INSERT INTO users (id, username, email, password) VALUES ($1, $2, $3, $4)"
	
	_, err := pool.Exec(
		t.Context(),
		sql,
		uuid.New(), "testuser", "testemail", "testpassword",
	)
	if err != nil {
		logger.Log.Fatal(
			"Не удалось создать тестового пользователя",
			zap.Error(err),
		)
	}
}

func SetUp(t *testing.T) *pgxpool.Pool {
	t.Helper()

	pool := CreatePool(t)
	makeMigrations(t, pool)
	createTestUser(t, pool)

	database.SetDBPool(pool)

	return pool
}