package chat

import (
	"log"
	"net/http"

	"github.com/andro-kes/Chat/shared/middlewares"
	"github.com/andro-kes/Chat/shared/models"

	"github.com/gin-gonic/gin"
)

func GetUserRooms(c *gin.Context) {
	currentUser := getCurrentUser(c)
	log.Println("Получение комнат для", currentUser.Username)
    
    var rooms []models.Room
    err := middlewares.DB.Model(&currentUser).
        Preload("Users").
        Preload("Admin").
        Association("Rooms").
        Find(&rooms)
        
    if err != nil {
		log.Printf("Не удалось получить комнаты из-за %s", err.Error())
        c.JSON(http.StatusInternalServerError, gin.H{
			"mesage": "Ошибка загрузки комнат",
			"error": err.Error(),
		})
        return
    }

    data := make(map[uint]string)
    for _, room := range rooms {
        data[room.ID] = room.Name
    }
	log.Println("Комнаты успешно получены")

    c.JSON(200, data)
}