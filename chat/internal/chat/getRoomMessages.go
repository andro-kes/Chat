package chat

import (
	"log"

	"github.com/andro-kes/Chat/shared/middlewares"
	"github.com/andro-kes/Chat/shared/models"

	"github.com/gin-gonic/gin"
)

func GetRoomMessages(c *gin.Context) {
	var messages []models.Message
	obj := middlewares.DB.Where("room_id = ?", c.Param("id")).
		Preload("Sender").
		Order("created_at desc").
		Find(&messages)

	if obj.Error != nil {
		c.JSON(200, "Сообщений нет")
		return
	}
	log.Println("Сообщения загружены")
	
	c.JSON(200, messages)
}