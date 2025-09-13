package services

import (
	"github.com/andro-kes/Chat/auth/internal/models"
	"github.com/andro-kes/Chat/auth/internal/repository"
	"github.com/andro-kes/Chat/auth/internal/utils"
	"github.com/andro-kes/Chat/auth/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserService interface {
	Login(user models.User) (*LoginData, error)
	Logout(token string) error
	SignUp()
	Update()
}

type userService struct {
	Repo         *repository.DBUserRepo
	TokenService TokenService
}

func NewUserService() *userService {
	return &userService{
		Repo: repository.NewUserRepo(),
		TokenService: NewTokenService(),
	}
}

type LoginData struct {
	User               *models.User
	RefreshTokenString string
	AccessTokenString  string
}

func (us *userService) Login(user models.User) (*LoginData, error) {
	logger.Log.Info(
		"Попытка входа в систему",
		zap.String("email", user.Email),
	)

	existingUser, err := us.Repo.FindByEmail(user.Email)
	if err != nil {
		return &LoginData{}, err
	}

	err = utils.CompareHashPasswords(existingUser.Password, user.Password)
	if err != nil {
		logger.Log.Warn(
			"Пароли не совпадают",
			zap.Error(err),
		)
		return &LoginData{}, err
	}

	refreshTokenString, err := us.TokenService.GenerateRefreshToken(existingUser.ID)
	if err != nil {
		return &LoginData{}, err
	}

	accessTokenString, err := us.TokenService.GenerateAccessToken(existingUser.ID)
	if err != nil {
		return &LoginData{}, err
	}

	return &LoginData{
		User:               existingUser,
		RefreshTokenString: refreshTokenString,
		AccessTokenString:  accessTokenString,
	}, nil
}

func (us userService) Logout(token string) error {
	tokenStringID, err := us.TokenService.ParseRefreshToken(token)
	if err != nil {
		return err
	}

	tokenID, err := uuid.Parse(tokenStringID)
	if err != nil {
		logger.Log.Error(
			"Неверный ID для refresh token",
			zap.Error(err),
		)
		return err
	}

	err = us.TokenService.(*tokenService).TokenRepo.DeleteByID(tokenID)
	
	return err
}

func (us userService) SignUp() {

}

func (us userService) Update() {

}
