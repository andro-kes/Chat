package services

import (
	"errors"
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
	RevokeRefreshToken(userID uuid.UUID) error
	ParseRefreshToken(tokenString string) (string, error) 
	ParseAccessToken(tokenString string) (string, error)
	ParseTokenClaims(tokenString string) (*jwt.RegisteredClaims, error)
	UpdateRefreshToken(refreshToken string) (string, error)
	UpdateAccessToken(accessToken string) (string, error)
}

type tokenService struct {
	TokenRepo repository.TokenRepo
}

func NewTokenService() *tokenService {
	return &tokenService{
		TokenRepo: repository.NewTokenRepo(),
	}
}

func (token *tokenService) GenerateRefreshToken(userId uuid.UUID) (string, error) {
	logger.Log.Info("Генерация нового refresh token")

	newRefreshTokenID := uuid.New()
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(720 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "auth",
		Subject:   userId.String(),
		ID:        newRefreshTokenID.String(),
		Audience:  []string{"auth", "chat"},
	}

	generatedRefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secretKey := []byte(os.Getenv("SECRET_KEY"))
	tokenString, err := generatedRefreshToken.SignedString(secretKey)
	if err != nil {
		logger.Log.Error("Не удалось сгенерировать refresh token", zap.Error(err))
		return "", err
	}

	// Сохраняем в репозитории: ассоциация user <-> refresh token
	if err := token.TokenRepo.Save(userId, newRefreshTokenID, tokenString); err != nil {
		logger.Log.Error("Не удалось сохранить refresh token в репозитории", zap.Error(err))
		return "", err
	}

	return tokenString, nil
}

func (token *tokenService) GenerateAccessToken(userId uuid.UUID) (string, error) {
	logger.Log.Info("Генерация нового access токена")

	newAccessTokenID := uuid.New()
	claims := jwt.RegisteredClaims{
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
		logger.Log.Error("Не удалось создать access token", zap.Error(err))
		return "", err
	}

	return tokenString, nil
}

// RevokeRefreshToken удаляет refresh-токены по userID (ревокация всех токенов пользователя).
func (token *tokenService) RevokeRefreshToken(userID uuid.UUID) error {
	return token.TokenRepo.DeleteByUserID(userID)
}

// ParseRefreshToken возвращает subject (user id) из refresh token.
// Комментарий: возвращаем user_id в виде строки (так удобнее далее парсить UUID).
func (token *tokenService) ParseRefreshToken(tokenString string) (string, error) {
	claims, err := token.ParseTokenClaims(tokenString)
	if err != nil {
		return "", err
	}
	return claims.Subject, nil
}

// ParseAccessToken возвращает subject (user id) из access token
func (token *tokenService) ParseAccessToken(tokenString string) (string, error) {
	claims, err := token.ParseTokenClaims(tokenString)
	if err != nil {
		return "", err
	}
	return claims.Subject, nil
}

// ParseTokenClaims парсит JWT и возвращает RegisteredClaims
func (token *tokenService) ParseTokenClaims(tokenString string) (*jwt.RegisteredClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(os.Getenv("SECRET_KEY")), nil
	})
	if err != nil {
		logger.Log.Error("Не удалось спарсить token claims", zap.Error(err))
		return nil, err
	}

	claims, ok := parsedToken.Claims.(*jwt.RegisteredClaims)
	if !ok {
		logger.Log.Error("Не удалось привести claims к RegisteredClaims")
		return nil, errors.New("неверный формат claims")
	}

	return claims, nil
}

// UpdateRefreshToken вращает (rotates) refresh token:
// - парсит переданный refresh token чтобы получить user_id,
// - создаёт новый refresh token для этого user_id,
// - обновляет запись в репозитории (по user_id) и возвращает новый token.
func (token *tokenService) UpdateRefreshToken(refreshToken string) (string, error) {
	claims, err := token.ParseTokenClaims(refreshToken)
	if err != nil {
		return "", err
	}

	userIdStr := claims.Subject
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		logger.Log.Warn("Невалидный user id в refresh token", zap.String("subject", userIdStr), zap.Error(err))
		return "", err
	}

	newToken, err := token.GenerateRefreshToken(userId)
	if err != nil {
		return "", err
	}

	// Обновляем запись по user_id; репозиторий должен реализовать эту операцию
	if err := token.TokenRepo.UpdateRefreshToken(userId, newToken); err != nil {
		logger.Log.Error("Не удалось обновить refresh token в репозитории", zap.Error(err))
		return "", err
	}

	return newToken, nil
}

// UpdateAccessToken парсит переданный access token чтобы получить user_id и возвращает новый access token.
func (token *tokenService) UpdateAccessToken(accessToken string) (string, error) {
	claims, err := token.ParseTokenClaims(accessToken)
	if err != nil {
		return "", err
	}

	userIdStr := claims.Subject
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		return "", err
	}

	return token.GenerateAccessToken(userId)
}