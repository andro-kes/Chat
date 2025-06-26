package chat

import (
	"github.com/andro-kes/Chat/shared/middlewares"
	"github.com/andro-kes/Chat/shared/models"
	"github.com/gin-gonic/gin"
)

type RoomName struct {
	Name string `json:"name"`
}

func CreateRoom(c *gin.Context) {
	currentUser := getCurrentUser(c)
	var roomName RoomName
	if err := c.ShouldBindBodyWithJSON(&roomName); err != nil {
		c.JSON(400, gin.H{"CreateRoom": "Неверные данные для создания комнаты"})
		return
	}
	NewRoom := models.Room{}
	NewRoom.AdminID = currentUser.ID
	NewRoom.Name = roomName.Name
	NewRoom.Users = []*models.User{&currentUser}

	obj := middlewares.DB.Create(&NewRoom)
	if obj.Error != nil {
		c.JSON(400, gin.H{"CreateRoom": obj.Error.Error()})
		return
	}

	c.JSON(200, gin.H{"CreateRoom": "Success", "room": NewRoom})
}