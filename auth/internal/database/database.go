// ВРЕМЕННО: Пакет database инкапсулирует создание и доступ к пулу соединений
// PostgreSQL через pgxpool. Комментарии временные, используются в рамках
// текущего рефакторинга.
package database

import (
	"context"
	"os"
	"time"

	"github.com/andro-kes/Chat/auth/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var dbPool *pgxpool.Pool

// Init ВРЕМЕННО: конфигурирует пул соединений на основании переменной окружения
// DB_USER_URL, задает базовые параметры пула и сохраняет ссылку в локальной
// переменной пакета для последующего использования.
func Init() {
	db_url := os.Getenv("DB_USER_URL")
	if db_url == "" {
		logger.Log.Panic(
			"Отсутствует ссылка на user_db",
			zap.String("db", "user_db"),
		)
	}
	
	config, err := pgxpool.ParseConfig(db_url)
	if err != nil {
		logger.Log.Panic(
			"Не удалось спарсить строку настроек db_user", 
			zap.Error(err),
		)
	}

	config.MaxConns = 50
	config.MinConns = 10
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		logger.Log.Panic(
			"Не удалось создать пул подключения",
			zap.String("db", "user_db"),
			zap.Error(err),
		)
	}

	SetDBPool(pool)
}

// GetDBPool ВРЕМЕННО: предоставляет доступ к глобальному пулу соединений
func GetDBPool() *pgxpool.Pool {
	return dbPool
}

func SetDBPool(pool *pgxpool.Pool) {
	dbPool = pool
}