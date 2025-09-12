package logger

import (
	"go.uber.org/zap"
)

var Log *zap.Logger

func Init() {
	config := zap.NewProductionConfig()
    config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)

    logger, err := config.Build()
    if err != nil {
        panic(err)
    }
	Log = logger
}

func Close() {
	Log.Sync()
}