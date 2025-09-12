package services

import (
	"github.com/andro-kes/Chat/auth/logger"
	"github.com/andro-kes/Chat/auth/internal/models"
	"github.com/andro-kes/Chat/auth/internal/repository"
	"github.com/andro-kes/Chat/auth/internal/utils"
	"go.uber.org/zap"
)


type UserServiceRepo interface {
	Login(user models.User) 
	Logout()
	SignUp()
	Update()
}

type UserService struct {
	Repo *repository.DBUserRepo
}

func NewUserService() *UserService {
	return &UserService{
		Repo: repository.NewUserRepo(),
	}
}

func (us *UserService) Login(user models.User) (*models.User, error) {
	logger.Log.Info(
		"Попытка входа в систему",
		zap.String("email", user.Email),
	)

	existingUser, err := us.Repo.FindByEmail(user.Email)
	if err != nil {
		return &models.User{}, err
	}

	err = utils.CompareHashPasswords(existingUser.Password, user.Password)
	if err != nil {
		logger.Log.Warn(
			"Пароли не совпадают",
			zap.Error(err),
		)
	}

	return existingUser, err
}

func (us *UserService) Logout() {
	
}

func (us *UserService) SignUp() {
	
}

func (us *UserService) Update() {
	
}