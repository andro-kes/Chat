package auth

import (
	"github.com/andro-kes/Chat/auth/internal/utils"
	"github.com/andro-kes/Chat/shared/middlewares"
	"github.com/andro-kes/Chat/shared/models"
	"github.com/gin-gonic/gin"
)

func UpdateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{
			"message": "Невалидные данные",
			"error": err.Error(),
		})
		return
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

	existingUser.Username = user.Username

	password, err := utils.GenerateHashPassword(user.Password)
	if err != nil {
		c.JSON(400, gin.H{
			"message": "Не удалось получить hash пароля",
			"error": err.Error(),
		})
		return
	}
	existingUser.Password = string(password)

	obj = middlewares.DB.Save(&existingUser)
	if obj.Error != nil {
		c.JSON(400, gin.H{
			"message": "Не удалось сохранить изменения",
			"error": obj.Error.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"message": "Данные успешно обновлены"})
}