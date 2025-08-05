package auth

import (
	"log"
	"net/http"
	"time"

	"github.com/andro-kes/Chat/auth/internal/utils"
	"github.com/andro-kes/Chat/shared/models"
	"github.com/gin-gonic/gin"
)

func SignUPPageHandler(c *gin.Context) {
	c.HTML(200, "signUp.html", nil)
}

func SignUpHandler(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Println("Не удалось зарегистрировать пользователя")
		c.JSON(400, gin.H{
			"message": "Не удалось зарегистрировать пользователя",
			"error": err.Error(),
		})
	}

	DB := utils.GetDB(c)

	var existingUser models.User
	DB.Where("email = ?", user.Email).First(&existingUser)

	if existingUser.ID != 0 {
		c.JSON(400, gin.H{
			"message": "Пользователь с таким email уже существует",
		})
		return
	}

	password, err := utils.GenerateHashPassword(user.Password)
	if err != nil {
		c.JSON(400, gin.H{
			"message": "Ошибка хэширования пароля",
			"error": err.Error(),
		})
	}
	user.Password = string(password)

	DB.Create(&user)
	log.Println("Создан новый пользователь", user.Username)

	user.Password = ""

	refreshToken, err := utils.GenerateRefreshToken(DB, user.ID)
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

	c.JSON(200, gin.H{"message": "Создан новый пользователь"})
}