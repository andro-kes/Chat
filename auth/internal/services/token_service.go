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

func (token *tokenService) GenerateRefreshToken(userId uuid.UUID) (string, error){
	logger.Log.Info(
		"Генерация нового refresh token",
	)

	newRefreshTokenID := uuid.New()
	claims := jwt.RegisteredClaims {
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(720 * time.Hour)),
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

func (token *tokenService) RevokeRefreshToken(userID uuid.UUID) error {
	return token.TokenRepo.DeleteByID(userID)
}

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

// Возвращает token_id string
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

	claims := parsedToken.Claims

	return claims.GetSubject()
}

// Возвращает id пользователя
func (token *tokenService) ParseAccessToken(tokenString string) (string, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(os.Getenv("SECRET_KEY")), nil
	})
	if err != nil {
		logger.Log.Error(
			"Не удалось спарсить access token",
			zap.Error(err),
		)
		return "", err
	}
	
	claims := parsedToken.Claims

	return claims.GetSubject()
}

// Возвращает claims токена
func (token *tokenService) ParseTokenClaims(tokenString string) (*jwt.RegisteredClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(os.Getenv("SECRET_KEY")), nil
	})
	if err != nil {
		logger.Log.Error(
			"Не удалось спарсить access token claims",
			zap.Error(err),
		)
		return nil, err
	}
	
	claims, ok := parsedToken.Claims.(*jwt.RegisteredClaims)
	if !ok {
		logger.Log.Error("Не удалось привести claims к RegisteredClaims")
		return nil, errors.New("неверный формат claims")
	}

	return claims, nil
}

func (token *tokenService) UpdateRefreshToken(refreshToken string) (string, error) {
	id, err := token.ParseRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}

	userId, err := uuid.Parse(id)
	if err != nil {
		logger.Log.Warn(
			"Невалидное id",
			zap.String("id", id),
		)
		return "", err
	}

	newToken, err := token.GenerateRefreshToken(userId)
	if err != nil {
		return "", err
	}

	return newToken, token.TokenRepo.UpdateRefreshToken(userId, newToken)
}

func (token *tokenService) UpdateAccessToken(accessToken string) (string, error) {
	id, err := token.ParseAccessToken(accessToken)
	if err != nil {
		return "", err
	}
	userId, err := uuid.Parse(id)
	if err != nil {
		return "", err
	}
	
	return token.GenerateAccessToken(userId)
}