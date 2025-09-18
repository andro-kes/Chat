// ВРЕМЕННО: Пакет configs отвечает за инициализацию глобальных конфигураций,
// таких как логгер и внешние подключения. Комментарии временные.
package configs

import "github.com/andro-kes/Chat/auth/logger"

// Init ВРЕМЕННО: базовая инициализация логирования приложения
func Init() {
	logger.Init()
	

	defer logger.Close()
}