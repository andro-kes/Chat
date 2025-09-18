// ВРЕМЕННО: Пакет services инкапсулирует логику генерации и валидации токенов
// (access/refresh) с использованием JWT и хранением refresh токенов в БД.
package services

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/andro-kes/Chat/auth/internal/repository"
	"github.com/andro-kes/Chat/auth/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TokenService interface {
	GenerateRefreshToken(userId uuid.UUID) (string, error)
	GenerateAccessToken(userId uuid.UUID) (string, error)
	RevokeRefreshToken(tokenID uuid.UUID)
	ParseRefreshToken(tokenString string) (string, error)
}

type tokenService struct {
	TokenRepo repository.TokenRepo
}

func NewTokenService() *tokenService {
	return &tokenService{
		TokenRepo: repository.NewTokenRepo(),
	}
}

// GenerateRefreshToken ВРЕМЕННО: выпускает refresh token и сохраняет его в БД
func (token *tokenService) GenerateRefreshToken(userId uuid.UUID) (string, error){
	logger.Log.Info(
		"Генерация нового refresh token",
	)

	newRefreshTokenID := uuid.New()
	claims := jwt.RegisteredClaims {
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()), // время создания
		NotBefore: jwt.NewNumericDate(time.Now()), // становится действительным
		Issuer:    "auth", // сервис издателя
		Subject:   userId.String(),
		ID:        newRefreshTokenID.String(),
		Audience:  []string{"auth", "chat"},
	}

	generatedRefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	secretKey := []byte(os.Getenv("SECRET_KEY"))
	tokenString, err := generatedRefreshToken.SignedString(secretKey)
	if err != nil {
		logger.Log.Error(
			"Не удалось сгенерировать refresh token",
			zap.Error(err),
		)
		return "", err
	}

	err = token.TokenRepo.Save(userId, newRefreshTokenID, tokenString)

	return tokenString, err
}

// GenerateAccessToken ВРЕМЕННО: выпускает короткоживущий access token
func (token *tokenService) GenerateAccessToken(userId uuid.UUID) (string, error) {
	logger.Log.Info(
		"Генерация нового access токена",
	)

	newAccessTokenID := uuid.New()
	claims := jwt.RegisteredClaims {
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		IssuedAt:  jwt.NewNumericDate(time.Now()), 
		NotBefore: jwt.NewNumericDate(time.Now()), 
		Issuer:    "auth",
		Subject:   userId.String(),
		ID:        newAccessTokenID.String(),
		Audience:  []string{"auth", "chat"},
	}

	generatedAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secretKey := []byte(os.Getenv("SECRET_KEY"))

	tokenString, err := generatedAccessToken.SignedString(secretKey)
	if err != nil {
		logger.Log.Error(
			"Не удалось создать access token",
			zap.Error(err),
		)
	}

	return tokenString, err
}

// RevokeRefreshToken ВРЕМЕННО: заглушка для отзыва refresh токена
func (token *tokenService) RevokeRefreshToken(tokenID uuid.UUID) {

}

// GetTokenCookie ВРЕМЕННО: утилита получения значения токена из cookie
func (token *tokenService) GetTokenCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		logger.Log.Warn(
			"Не удалось получить токен из cookie",
			zap.Error(err),
		)
		return "", err
	}

	return cookie.Value, nil
}

// ParseRefreshToken ВРЕМЕННО: валидирует и извлекает token_id из refresh токена
func (token *tokenService) ParseRefreshToken(tokenString string) (string, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(os.Getenv("SECRET_KEY")), nil
	})
	if err != nil {
		logger.Log.Error(
			"Не удалось спарсить refresh token",
			zap.Error(err),
		)
		return "", err
	}

	claims, ok := parsedToken.Claims.(jwt.RegisteredClaims)
	if !ok {
		logger.Log.Warn(
			"Невалидные claims для refresh token",
		)
		return "", errors.New("Невалидные claims")
	}

	return claims.ID, nil
}