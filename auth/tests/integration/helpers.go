package auth_tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/andro-kes/Chat/auth/internal/database"
	"github.com/andro-kes/Chat/auth/internal/handlers"
	"github.com/andro-kes/Chat/auth/internal/utils"
	"github.com/andro-kes/Chat/auth/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	 _ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

func createTestPool(t *testing.T) *pgxpool.Pool {
    ctx := context.Background()
    
    req := testcontainers.ContainerRequest{
        Image:        "postgres:15-alpine",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_USER":     "testuser",
            "POSTGRES_PASSWORD": "testpass",
            "POSTGRES_DB":       "testdb",
        },
        WaitingFor: wait.ForLog("database system is ready to accept connections").WithStartupTimeout(120 * time.Second),
    }
    
    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    if err != nil {
        logger.Log.Fatal(
			"Не удалось создать тестовый контейнер",
			zap.Error(err),
		)
    }
	logger.Log.Info("Тестовый контейнер успешно создан")
    
    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "5432/tcp")
    
    dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())
    
    pool, err := pgxpool.New(ctx, dsn)
    if err != nil {
        logger.Log.Fatal(
			"Не удалось создать пул соединения",
			zap.Error(err),
		)
    }

    // Даем Postgres время на полную готовность: несколько попыток Ping с бэкофом
    var pingErr error
    for attempt := 0; attempt < 20; attempt++ {
        pingErr = pool.Ping(ctx)
        if pingErr == nil {
            break
        }
        time.Sleep(time.Duration(250*(attempt+1)) * time.Millisecond)
    }
    if pingErr != nil {
        logger.Log.Fatal(
            "Не удалось подсоединиться к пулу",
            zap.Error(pingErr),
        )
    }
    
    t.Cleanup(func() {
        pool.Close()
        container.Terminate(ctx)
    })

	logger.Log.Info(
		"Пул соединения создан",
	)
    
    return pool
}

func makeMigrations(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

    fixtures := []string{
        `CREATE EXTENSION IF NOT EXISTS pgcrypto;`,
        `CREATE TABLE IF NOT EXISTS users (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP,
            deleted_at TIMESTAMP,
            username VARCHAR(255) NOT NULL,
            email VARCHAR(255) UNIQUE NOT NULL,
            password VARCHAR(255) NOT NULL
        );`,
        `CREATE TABLE IF NOT EXISTS refresh_tokens (
            user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            token_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            token TEXT NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT NOW()
        );`,
        `CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`,
        `CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);`,
        `CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_id ON refresh_tokens(token_id);`,
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

	logger.Log.Info(
		"Миграции загружены",
	)
}

func createTestUser(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	// Хешируем пароль, чтобы пройти сравнение bcrypt в сервисе
	passwordHash, err := utils.GenerateHashPassword("testpassword")
	if err != nil {
		logger.Log.Fatal(
			"Не удалось хэшировать пароль",
			zap.Error(err),
		)
	}

	sql := "INSERT INTO users (id, created_at, username, email, password) VALUES ($1, NOW(), $2, $3, $4)"
	_, err = pool.Exec(
		t.Context(),
		sql,
		uuid.New(), "testuser", "testemail", string(passwordHash),
	)
	if err != nil {
		logger.Log.Fatal(
			"Не удалось создать тестового пользователя",
			zap.Error(err),
		)
	}

	logger.Log.Info(
		"Тестовый пользователь создан",
	)
}

func SetUp(t *testing.T) *handlers.AuthHandlers {
	t.Helper()

	logger.Init()
	defer logger.Close()

	pool := createTestPool(t)
	makeMigrations(t, pool)
	createTestUser(t, pool)

	database.SetDBPool(pool)

	return handlers.NewAuthHandlers()
}