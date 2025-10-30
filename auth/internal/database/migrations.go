package database

import (
	"context"

	"github.com/andro-kes/Chat/auth/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func makeMigrations(ctx context.Context, pool *pgxpool.Pool) {
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