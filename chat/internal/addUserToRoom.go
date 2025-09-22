package chat

import (
	"github.com/andro-kes/Chat/shared/middlewares"
	"github.com/andro-kes/Chat/shared/models"
	"github.com/gin-gonic/gin"
)

type IDUser struct {
	ID uint `json:"id"`
}

func AddUserToRoom(c *gin.Context) {
	id := c.Param("id")
	currentRoom, err := getCurrentRoom(id)
	if err != nil {
		c.JSON(400, gin.H{"AddUserToRoom": "Не удалось получить доступ к комнате", "id_room": id})
		return
	}

	var idUser IDUser
	if err := c.ShouldBindBodyWithJSON(&idUser); err != nil {
		c.JSON(400, gin.H{"AddUserToRoom": "Некорректные данные"})
		return
	}

	var user models.User
	middlewares.DB.Where("id = ?", idUser.ID).First(&user)

	if err := middlewares.DB.Model(&currentRoom).Association("Users").Append(&user); err != nil {
		c.JSON(400, gin.H{"AddUserToRoom": "Не удалось добавить пользователя в комнату"})
		return
	}

	c.JSON(200, gin.H{"AddUserToRoom": "Success", "user": user})
}