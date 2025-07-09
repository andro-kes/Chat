package auth

import (
	"log"

	"github.com/andro-kes/Chat/auth/internal/utils"
	"github.com/andro-kes/Chat/shared/middlewares"
	"github.com/andro-kes/Chat/shared/models"
	"github.com/gin-gonic/gin"
)

func LogoutHandler(c *gin.Context) {
	DB := middlewares.DB

	token, err := c.Cookie("refresh_token")
	if err != nil {
		log.Println("LogoutHandler: Refresh токен не найден в cookie")
		c.JSON(400, gin.H{"error": "Refresh токен не найден"})
	}
	
	claims, err := utils.ParseRefreshToken(token)
	if err != nil {
		log.Println("LogoutHandler: Claims не определены")
		c.JSON(400, gin.H{"error": "Права пользователя не определены"})
	}

	var RefreshToken models.RefreshTokens
	obj := DB.Where("user_id = ?", claims.UserID).Delete(&RefreshToken)
	if obj.Error != nil {
		log.Println("LogoutHandler: Refresh токен не найден в базе")
		c.JSON(400, gin.H{"error": "Ошибка при получении refresh еокена из базы данных"})
	}

	c.SetCookie("refresh_token", "", -1, "/", "localhost", false, true)
	c.SetCookie("token", "", -1, "/", "localhost", false, true)
	c.JSON(200, gin.H{"message": "Пользователь вышел из системы"})
}