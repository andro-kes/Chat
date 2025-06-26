package middlewares

import (
	"log"
	"os"

	"github.com/andro-kes/Chat/shared/models"
	"github.com/andro-kes/Chat/shared/helpers"

	"github.com/gin-gonic/gin"
)

var currentUser models.User

func getCurrentUser(refreshToken string) models.User {
	var user models.User
	claims, err := helpers.ParseRefreshToken(refreshToken, os.Getenv("SECRET_KEY"))
	if err != nil {
		log.Printf("getCurrentUser: Не удалось извлечь claims\n%s", err.Error())
		user.ID = 0
		return user
	}
	
	obj := DB.Where("id = ?", claims.UserID).First(&user)
	if obj.Error != nil {
		log.Printf("getCurrentUser: Ошибка при попытке получить пользователя\n%s", obj.Error.Error())
		user.ID = 0
		return user
	}
	return user
}

func IsAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Лучше использовать access
		refreshToken, err := c.Cookie("refresh_token")
		if err != nil {
			log.Println("IsAuthMiddleware: не найден реферш токен в куках пользователя")
			c.JSON(400, gin.H{"Error": "Выполните вход"})
			return
		}
		currentUser = getCurrentUser(refreshToken)
		if currentUser.ID == 0 {
			return
		}
		c.Set("User", currentUser)
		c.Next()
	}
}