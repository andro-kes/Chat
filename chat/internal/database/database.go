// database.go для chat сервиса
package database

import (
	"context"
	"os"
	"time"

	"github.com/andro-kes/Chat/chat/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var dbPool *pgxpool.Pool

func Init() {
	db_url := os.Getenv("DB_CHAT_URL") // Используем DB_CHAT_URL
	if db_url == "" {
		logger.Log.Panic(
			"Отсутствует ссылка на chat_db",
			zap.String("db", "chat_db"),
		)
	}
	
	config, err := pgxpool.ParseConfig(db_url)
	if err != nil {
		logger.Log.Panic(
			"Не удалось спарсить строку настроек chat_db", 
			zap.Error(err),
		)
	}

	config.MaxConns = 50
	config.MinConns = 10
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = time.Minute

	ctx := context.Background()

	// Добавляем retry логику
	pool, err := connectWithRetry(ctx, config)
	if err != nil {
		logger.Log.Panic(
			"Не удалось создать пул подключения к chat_db",
			zap.String("db", "chat_db"),
			zap.Error(err),
		)
	}

	makeMigrations(ctx, pool)

	SetDBPool(pool)
	
	logger.Log.Info(
		"Успешно подключились к chat_db",
		zap.String("db", "chat_db"),
	)
}

func connectWithRetry(ctx context.Context, config *pgxpool.Config) (*pgxpool.Pool, error) {
	var pool *pgxpool.Pool
	var err error
	
	maxRetries := 10 
	retryDelay := time.Second * 3

	for i := range maxRetries {
		pool, err = pgxpool.NewWithConfig(ctx, config)
		if err == nil {
			// Проверяем соединение
			if pingErr := pool.Ping(ctx); pingErr == nil {
				return pool, nil
			}
			pool.Close()
		}
		
		if i < maxRetries-1 {
			logger.Log.Warn(
				"Не удалось подключиться к chat_db, повторная попытка",
				zap.Int("attempt", i+1),
				zap.Int("max_attempts", maxRetries),
				zap.Duration("next_attempt_in", retryDelay),
				zap.Error(err),
			)
			time.Sleep(retryDelay)
			retryDelay = time.Duration(float64(retryDelay) * 1.5)
		}
	}
	
	return nil, err
}

func GetDBPool() *pgxpool.Pool {
	return dbPool
}

func SetDBPool(pool *pgxpool.Pool) {
	dbPool = pool
}

func ClosePool() {
	if dbPool != nil {
		dbPool.Close()
	}
}