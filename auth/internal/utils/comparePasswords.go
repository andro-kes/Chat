package utils

import (
	"golang.org/x/crypto/bcrypt"
)

func CompareHashPasswords(userPassword, existingPassword string) error {
	err :=  bcrypt.CompareHashAndPassword([]byte(existingPassword), []byte(userPassword))
	return err
}