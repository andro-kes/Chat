package chat

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/andro-kes/Chat/shared/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func ChatHandler(c *gin.Context) {
	upgrader := websocket.Upgrader{
		CheckOrigin:     func(r *http.Request) bool { return true },
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

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

	conn.SetPingHandler(func(string) error {
        conn.SetReadDeadline(time.Now().Add(60 * time.Second))
        return conn.WriteControl(websocket.PongMessage, nil, time.Now().Add(5*time.Second))
    })

	for {
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