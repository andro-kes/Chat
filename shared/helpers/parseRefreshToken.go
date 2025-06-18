package helpers

import (
	"github.com/andro-kes/Chat/shared/models"
	"github.com/dgrijalva/jwt-go"
)

func ParseRefreshToken(tokenString string, secretKey string) (*models.RefreshTokensClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.RefreshTokensClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*models.RefreshTokensClaims)
	if !ok {
		return nil, err
	}

	return claims, err
}