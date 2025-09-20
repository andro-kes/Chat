// ВРЕМЕННО: Пакет services содержит бизнес-логику. UserService инкапсулирует
// операции аутентификации, входа по OAuth, управление токенами и паролями.
// Комментарии временные и помогут навигации до завершения рефакторинга.
package services

import (
	"errors"

	"github.com/andro-kes/Chat/auth/internal/models"
	"github.com/andro-kes/Chat/auth/internal/repository"
	"github.com/andro-kes/Chat/auth/internal/utils"
	"github.com/andro-kes/Chat/auth/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserService interface {
	Login(user *models.User) (*LoginData, error)
	OAuthLogin(username, email string) (*LoginData, error)
	Logout(token string) error
	SignUp(user *models.User) (*LoginData, error)
	SetPassword(user *models.User) error 
}

type userService struct {
	Repo         repository.UserRepo
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

// Login ВРЕМЕННО: валидирует пользователя, сравнивает пароли и выдает
// refresh/access токены
func (us *userService) Login(user *models.User) (*LoginData, error) {
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

// OAuthLogin ВРЕМЕННО: логика входа/регистрации через OAuth. Если пользователя
// нет — создается, иначе выполняется стандартный вход. Возвращает LoginData.
func (us *userService) OAuthLogin(username, email string) (*LoginData, error) {
	user, err := us.Repo.FindByEmail(email)
	if err != nil {
		newUser := &models.User{
			Username: username,
			Email: email,
		}
		us.Repo.CreateUser(newUser)
		return &LoginData{User: newUser}, errors.New("Пользователь был создан")
	}

	loginData, err := us.Login(user)
	if err != nil {
		logger.Log.Warn(
			"Пользователь не смог войти через oath",
			zap.Error(err),
		)
		return &LoginData{}, err
	}

	return loginData, nil
}

// Logout ВРЕМЕННО: парсит refresh token, получает его uuid и отзывает в хранилище
func (us *userService) Logout(token string) error {
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

// SignUp ВРЕМЕННО: заглушка под регистрацию пользователя
func (us *userService) SignUp(user *models.User) (*LoginData, error) {
	_, err := us.Repo.FindByEmail(user.Email)
	if err == nil {
		return &LoginData{}, errors.New("Пользователь с таким email уже существует")
	}

	hashPassword, err := utils.GenerateHashPassword(user.Password)
	if err != nil {
		logger.Log.Error(
			"Не удалось хэшировать пароль",
			zap.Error(err),
		)
		return &LoginData{}, err
	}
	password := user.Password
	user.Password = string(hashPassword)

	err = us.Repo.CreateUser(user)
	if err != nil {
		logger.Log.Warn(
			"Не удалось создать пользователя",
			zap.Error(err),
		)
	}

	user.Password = password

	return us.Login(user)
}

// SetPassword ВРЕМЕННО: устанавливает пароль через репозиторий
func (us *userService) SetPassword(user *models.User) error {
	return us.Repo.SetPassword(user)
}