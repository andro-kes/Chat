package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// Сравнивает hash пароль с введенным
func CompareHashPasswords(existingPassword, userPassword string) error {
	err :=  bcrypt.CompareHashAndPassword([]byte(existingPassword), []byte(userPassword))
	return err
}