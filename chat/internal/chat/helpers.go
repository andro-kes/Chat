package chat

import (
	"log"

	"github.com/andro-kes/Chat/shared/middlewares"
	"github.com/andro-kes/Chat/shared/models"
	"github.com/gin-gonic/gin"
)

func CheckAccess(room *models.Room, user *models.User) bool {
	var cnt int64
	middlewares.DB.Table("room_users").
		Where("room_id = ? AND user_id = ?", room.ID, user.ID).
		Count(&cnt)
	log.Println("Найдено совпадений", cnt)
	return cnt > 0
}

func getCurrentRoom(roomID string) (models.Room, error) {
	var currentRoom models.Room
	obj := middlewares.DB.Where("id = ?", roomID).First(&currentRoom)
	return currentRoom, obj.Error
}

func getCurrentUser(c *gin.Context) models.User {
	user, ok := c.Get("User")
	if !ok {
		log.Println("MainPage: контекст не содержит данных о пользователе")
		return models.User{}
	}

	currentUser, ok := user.(models.User)
	if !ok {
		return models.User{}
	}

	return currentUser
}