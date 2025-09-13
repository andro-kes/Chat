package configs

import "github.com/andro-kes/Chat/auth/logger"

func Init() {
	logger.Init()
	

	defer logger.Close()
}