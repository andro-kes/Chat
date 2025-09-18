// ВРЕМЕННО: Утилита генерации хеша пароля. Комментарии временные.
package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// GenerateHashPassword ВРЕМЕННО: создает bcrypt-хеш для заданного пароля
func GenerateHashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), 16)
}