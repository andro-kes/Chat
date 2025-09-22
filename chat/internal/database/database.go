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
	db_url := os.Getenv("DB_CHAT_URL")
	if db_url == "" {
		logger.Log.Panic(
			"Отсутствует ссылка на chat_db",
			zap.String("db", "chat_db"),
		)
	}
	
	config, err := pgxpool.ParseConfig(db_url)
	if err != nil {
		logger.Log.Panic(
			"Не удалось спарсить строку настроек db_chat", 
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
			zap.String("db", "chat_db"),
			zap.Error(err),
		)
	}

	SetDBPool(pool)
}

func GetDBPool() *pgxpool.Pool {
	return dbPool
}

func SetDBPool(pool *pgxpool.Pool) {
	dbPool = pool
}