package chat

import (
	"log"
	"strconv"
	"time"

	"github.com/andro-kes/Chat/shared/models"

	"github.com/gin-gonic/gin"
)

func ChatHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(400, gin.H{
			"ChatHandler": "Не удалось установить соединение с сокетом",
			"error": err.Error(),
		})
		return
	}
	log.Println("Соединение установлено")

	id := c.Param("id")
	roomID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(500, gin.H{"Error": "Неверный формат ID комнаты"})
		return
	}

	var (
		currentRoom *models.RoomData
		currentUser models.User
		message models.Message
		ok bool
	)

	if currentRoom, ok = Manager.ActiveRooms[uint(roomID)]; !ok {
		c.JSON(400, gin.H{"Error": "Нельзя отправлять сообщения в неактивную комнату"})
		return
	}

	if currentUser = getCurrentUser(c); currentUser.ID == 0 {
		c.JSON(400, gin.H{"Error": "Пользователь не найден"})
		return
	}

	currentUserData := &models.UserData{
		User: currentUser,
		Conn: conn,
	}
	currentRoom.Registered(currentUserData)

	defer func() {
		log.Println("Соединение разорвано")
		
		if currentRoom != nil {
			currentRoom.Unregistered(currentUserData)
		}

		if currentRoom != nil && !currentRoom.CheckActive() {
			Manager.Mu.Lock()
			defer Manager.Mu.Unlock()
			
			if _, ok := Manager.ActiveRooms[uint(roomID)]; ok {
				currentRoom.Stop()
				delete(Manager.ActiveRooms, uint(roomID))
			}
		}
	}()

	for {
		conn.SetReadLimit(1024)
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

		err := conn.ReadJSON(&message)
		if err != nil {
			log.Println("Ошибка считывания сообщения", err.Error())
			break
		}

		message.RoomID = currentRoom.Room.ID
		message.SenderID = currentUser.ID
		currentRoom.SendMessage(&message)
	}
}