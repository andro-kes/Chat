package utils

import (
	"time"
	"strconv"
	"os"
	"log"

	"github.com/andro-kes/Chat/shared/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
)

func GenerateAccessToken(existingUser models.User) (string, error) {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("GenerateAccessToken: Не удалось загрузить env файл")
	}
	expititionTime := time.Now().Add(5 * time.Minute)
	claims := models.Claims{
		StandardClaims: jwt.StandardClaims{
			Subject: strconv.Itoa(int(existingUser.ID)),
			ExpiresAt: expititionTime.Unix(),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := jwtToken.SignedString([]byte(os.Getenv("SECRET_KEY")))
	return tokenString, err
}