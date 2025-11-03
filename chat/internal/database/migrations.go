package database

import (
	"context"
	"time"

	"github.com/andro-kes/Chat/chat/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func makeMigrations(ctx context.Context, pool *pgxpool.Pool) {
    fixtures := []string{
        `CREATE EXTENSION IF NOT EXISTS pgcrypto;`,
        `CREATE TABLE IF NOT EXISTS rooms (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            name VARCHAR(255) NOT NULL,
            created_by UUID NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP
        );`,
        `CREATE TABLE IF NOT EXISTS messages (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
            user_id UUID NOT NULL,
            content TEXT NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT NOW()
        );`,
        `CREATE TABLE IF NOT EXISTS room_users (
            room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
            user_id UUID NOT NULL,
            joined_at TIMESTAMP NOT NULL DEFAULT NOW(),
            PRIMARY KEY (room_id, user_id)
        );`,
        `CREATE INDEX IF NOT EXISTS idx_messages_room_id ON messages(room_id);`,
        `CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);`,
        `CREATE INDEX IF NOT EXISTS idx_rooms_created_by ON rooms(created_by);`,
    }

    // Добавьте retry логику для миграций...
	for _, fixture := range fixtures {
		var err error
		maxRetries := 3
		for attempt := range maxRetries {
			_, err = pool.Exec(ctx, fixture)
			if err == nil {
				break
			}
			if attempt < maxRetries-1 {
				logger.Log.Warn(
					"Не удалось выполнить миграцию для chat_db, повторная попытка",
					zap.Int("attempt", attempt+1),
					zap.Error(err),
				)
				time.Sleep(2 * time.Second)
			}
		}
		
		if err != nil {
			logger.Log.Fatal(
				"Не удалось провести миграцию для chat_db",
				zap.Error(err),
			)
		}
	}

	logger.Log.Info(
		"Миграции для chat_db загружены",
	)
}