package chat

import (
	"log"

	"github.com/andro-kes/Chat/shared/models"
	"github.com/andro-kes/Chat/shared/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader websocket.Upgrader

func MainPageHandler(c *gin.Context) {
	user, ok := c.Get("User")
	if !ok {
		log.Println("MainPage: контекст не содержит данных о пользователе")
		return
	}

	currentUser, ok := user.(models.User)
	if !ok {
		return
	}

	c.HTML(200, "main.html", currentUser.Rooms)
}


func ChatPageHandler(c *gin.Context) {
	roomID := c.Param("id")

	var room models.Room
	obj := middlewares.DB.Where("id = ?", roomID).First(&room)
	if obj.Error != nil {
		log.Printf("ChatPage: Не удалось получить комнату\n%s", obj.Error.Error())
		c.JSON(404, gin.H{"Error": "Комната не найдена"})
		return
	}

	currentRoom := &models.RoomData{
		Room: room,
		Broadcast: make(chan *models.Message),
		Registered: make(chan *models.UserData),
		Unregistered: make(chan *models.UserData),
	}

	go func(room *models.RoomData) {
		room.Run()
	}(currentRoom)

	c.HTML(200, "chat.html", nil)
}


func ChatHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Не удалось установить соединение")
		c.JSON(400, gin.H{"Error": "Не удалось установить соединение с сокетом"})
		return
	}

	var message models.Message

	for {
		defer func() {
			// Отправлять в канал Unregistered
			conn.Close()
		} ()

		err := conn.ReadJSON(&message)
		if err != nil {
			log.Println("Ошибка считывания сообщения")
			break
		}

		err = conn.WriteJSON(message)
		if err != nil {
			log.Println("Ошибка отправления сообщения")
			break
		}
	}
}