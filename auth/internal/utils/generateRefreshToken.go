package utils

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

func GenerateRefreshToken(DB *gorm.DB, userID uint) (string, error) {
	log.Println("Генерация рефреш токена")
	expititionTime := time.Now().Add(7 * 24 * time.Hour)

	TokenID := uuid.New()

	refreshClaims := models.RefreshTokensClaims{
		UserID: userID,
		TokenID: TokenID,
		StandardClaims: jwt.StandardClaims{
			Subject: strconv.Itoa(int(userID)),
			ExpiresAt: expititionTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
	
	var existingToken models.RefreshTokens
	DB.Where("user_id = ?", userID).First(&existingToken)
	if existingToken.Token != "" {
		log.Println("Обновление реферш токен")
		DB.Model(&existingToken).Update("token", tokenString)
		return tokenString, err
	}

	refreshToken := &models.RefreshTokens{
		UserID: userID,
		TokenID: TokenID,
		Token: tokenString,
	}

	log.Println("Добавления рефреш токена в базу", refreshToken.TokenID)
	obj := DB.Create(&refreshToken)
	if obj.Error != nil {
		log.Println("Ошибка при добавлении рефреш токена в базу")
		return "", obj.Error
	}

	return tokenString, err
}