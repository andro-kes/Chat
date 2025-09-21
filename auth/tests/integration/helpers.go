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
	"github.com/docker/go-connections/nat"
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
			"SECRET_KEY":        "secretkey",
        },
        WaitingFor: wait.ForSQL("5432/tcp", "pgx", func(host string, port nat.Port) string {
            return fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())
        }).WithStartupTimeout(120 * time.Second),
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
    
    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "5432")
    
    dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb", host, port.Port())
    
    pool, err := pgxpool.New(ctx, dsn)
    if err != nil {
        logger.Log.Fatal(
			"Не удалось создать пул соединения",
			zap.Error(err),
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
		`CREATE TABLE users(
			id UUID PRIMARY KEY,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP,
			deleted_at TIMESTAMP,
			username VARCHAR(255),
			email VARCHAR(255),
			password VARCHAR(255)
		)
		`,
		`CREATE TABLE refresh_tokens(
			user_id UUID NOT NULL,
			token_id UUID PRIMARY KEY,
			token VARCHAR(500) NOT NULL
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

	logger.Log.Info(
		"Миграции загружены",
	)
}

func createTestUser(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	// Хешируем пароль для тестового пользователя
	hashedPassword, err := utils.GenerateHashPassword("testpassword")
	if err != nil {
		logger.Log.Fatal(
			"Не удалось хешировать пароль для тестового пользователя",
			zap.Error(err),
		)
	}

	sql := "INSERT INTO users (id, username, email, password) VALUES ($1, $2, $3, $4)"
	
	_, err = pool.Exec(
		t.Context(),
		sql,
		uuid.New(), "testuser", "testemail", string(hashedPassword),
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