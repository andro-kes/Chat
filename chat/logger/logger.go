// ВРЕМЕННО: Тонкая обертка над zap-логгером для централизованного логирования.
package logger

import (
	"go.uber.org/zap"
)

var Log *zap.Logger

// Init ВРЕМЕННО: инициализирует глобальный zap-логгер в production-конфигурации
func Init() {
	config := zap.NewProductionConfig()
    config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)

    logger, err := config.Build()
    if err != nil {
        panic(err)
    }
	Log = logger
}

// Close ВРЕМЕННО: корректно синхронизирует и завершает работу логгера
func Close() {
	Log.Sync()
}