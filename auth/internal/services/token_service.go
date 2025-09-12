package services

import (
	"net/http"
	"os"
	"time"

	"github.com/andro-kes/Chat/auth/internal/repository"
	"github.com/andro-kes/Chat/auth/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TokenServiceRepo interface {
	GenerateRefreshToken(userId uuid.UUID) (string, error)
	GenerateAccessToken(userId uuid.UUID) (string, error)
	SetTokenCookie(w http.ResponseWriter, name, value string, expires time.Time)
}

type TokenService struct {
	tokenRepo *repository.DBTokenRepo
}

func NewTokenService() *TokenService {
	return &TokenService{
		tokenRepo: repository.NewTokenRepo(),
	}
}

func (token *TokenService) GenerateRefreshToken(userId uuid.UUID) (string, error){
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

	err = token.tokenRepo.Save(userId, newRefreshTokenID, tokenString)

	return tokenString, err
}

func (token *TokenService) GenerateAccessToken(userId uuid.UUID) (string, error) {
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

	return tokenString, err
}

func (token *TokenService) SetTokenCookie(w http.ResponseWriter, name, value string) {
	cookie := &http.Cookie{
        Name:     name,
        Value:    value,
        Path:     "/",
        HttpOnly: true, // Доступ только через HTTP, защита от XSS
        Secure:   true, // Только HTTPS
        SameSite: http.SameSiteStrictMode, // Защита от CSRF
    }
    http.SetCookie(w, cookie)
}