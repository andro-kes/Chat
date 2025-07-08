package chat

import (
	"log"
	"strconv"

	"github.com/andro-kes/Chat/shared/models"

	"github.com/gin-gonic/gin"
)

const WORKERS_COUNT = 5

func ChatPageHandler(c *gin.Context) {
	id := c.Param("id")
	roomID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(500, gin.H{
			"ChatPageHandler": "Неверный формат ID комнаты",
			"error": err.Error(),
		})
		return
	}

	currentRoom, err := getCurrentRoom(id)
	if err != nil {
		c.JSON(404, gin.H{
			"ChatPageHandler": "Такой комнаты не существует",
			"error": err.Error(),
		})
		return
	}
	log.Println("Вход в комнату:", currentRoom.Name)

	currentUser := getCurrentUser(c)
	if !CheckAccess(&currentRoom, &currentUser) {
		c.JSON(404, gin.H{
			"ChatPageHandler": "Доступ запрещен",
			"User": currentUser.Username,
		})
		return
	}

	if _, ok := Manager.ActiveRooms[uint(roomID)]; !ok {
		newRoom := &models.RoomData{
			Room: currentRoom,
			ActiveUsers: make(map[uint]*models.UserData),
		}
		newRoom.StartWork(WORKERS_COUNT)
		Manager.AddRoom(newRoom)
	}

	c.HTML(200, "chat.html", nil)
}