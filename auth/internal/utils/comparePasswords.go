// ВРЕМЕННО: Пакет utils содержит вспомогательные утилиты, включая работу с
// паролями. Комментарии временные и будут уточняться при рефакторинге.
package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// CompareHashPasswords ВРЕМЕННО: сравнивает хеш пароля с введенным паролем
func CompareHashPasswords(existingPassword, userPassword string) error {
	err :=  bcrypt.CompareHashAndPassword([]byte(existingPassword), []byte(userPassword))
	return err
}