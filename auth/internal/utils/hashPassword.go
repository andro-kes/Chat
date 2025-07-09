package utils

import (
	"golang.org/x/crypto/bcrypt"
)

func GenerateHashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), 16)
}