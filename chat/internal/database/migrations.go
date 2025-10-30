package database

import (
	"context"

	"github.com/andro-kes/Chat/chat/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// makeMigrations выполняет миграции базы данных.
// Создает таблицы и индексы, необходимые для работы сервиса.
func makeMigrations(ctx context.Context, pool *pgxpool.Pool) {
	fixtures := []string{
		`CREATE EXTENSION IF NOT EXISTS pgcrypto;`,

		`CREATE TABLE IF NOT EXISTS rooms (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			admin_id UUID NOT NULL,
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP,
			deleted_at TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_rooms_admin_id ON rooms(admin_id);`,
		`CREATE INDEX IF NOT EXISTS idx_rooms_created_at ON rooms(created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_rooms_deleted_at ON rooms(deleted_at);`,

		`CREATE TABLE IF NOT EXISTS messages (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			sender_id UUID NOT NULL,
			room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);`,
		`CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages(sender_id);`,
		`CREATE INDEX IF NOT EXISTS idx_messages_room_id ON messages(room_id);`,
		`CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);`,
	}

	for _, fixture := range fixtures {
		_, err := pool.Exec(ctx, fixture)
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
