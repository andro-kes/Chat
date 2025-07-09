package auth

import (
	"net/http"
	"time"

	"github.com/andro-kes/Chat/auth/internal/utils"
	"github.com/andro-kes/Chat/shared/middlewares"
	"github.com/andro-kes/Chat/shared/models"
	"github.com/gin-gonic/gin"
)

func LoginPageHandler(c *gin.Context) {
	c.HTML(200, "login.html", nil)
}

func Login(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{
			"message": "Невалидные данные",
			"error": err.Error(),
		})
	}

	var existingUser models.User
	obj := middlewares.DB.Where("email = ?", user.Email).First(&existingUser)
	if obj.Error != nil {
		c.JSON(400, gin.H{
			"message": "Пользователь не найден",
			"error": obj.Error.Error(),
		})
		return
	}

	err := utils.CompareHashPasswords(user.Password, existingUser.Password)
	if err != nil {
		c.JSON(400, gin.H{
			"message": "Пароли не совпадают",
			"error": err.Error(),
		})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(middlewares.DB, existingUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tokenString, err := utils.GenerateAccessToken(existingUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка при создании токена"})
		return
	}
	expititionTime := time.Now().Add(5 * time.Minute)

	c.SetCookie("refresh_token", refreshToken, int(time.Now().Add(7*24*time.Hour).Unix()), "/", "localhost", false, true)
	c.SetCookie("token", tokenString, int(expititionTime.Unix()), "/", "localhost", false, true)

	c.JSON(200, gin.H{"message": "Пользователь вошел в систему"})
}