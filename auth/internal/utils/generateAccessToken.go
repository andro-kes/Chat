package utils

import (
	"time"
	"strconv"
	"os"

	"github.com/andro-kes/Chat/shared/models"

	"github.com/dgrijalva/jwt-go"
)

func GenerateAccessToken(existingUser models.User) (string, error) {
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