package chat

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/andro-kes/Chat/shared/middlewares"
	"github.com/andro-kes/Chat/shared/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Пока что локально доверяю всем серверам
        return true
    },
}

var Manager = RoomManager{
	ActiveRooms: make(map[uint]*models.RoomData),
}

func MainPageHandler(c *gin.Context) {
	currentUser := getCurrentUser(c)
	if currentUser.ID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"Error": "Пользователь не авторизован"})
		return
	}

	data := make(map[uint]string)
	for _, room := range currentUser.Rooms {
		data[room.ID] = room.Name
	}

	c.HTML(200, "main.html", data)
}


func ChatPageHandler(c *gin.Context) {
	id := c.Param("id")
	roomID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(500, gin.H{"Error": "Неверный формат ID комнаты"})
		return
	}

	currentRoom, err := getCurrentRoom(id)
	if err != nil {
		c.JSON(404, gin.H{"ChatPageHandler": "Такой комнаты не существует"})
		return
	}

	currentUser := getCurrentUser(c)
	if !currentRoom.CheckAccess(&currentUser) {
		c.JSON(404, gin.H{"Error": "Доступ запрещен"})
		return
	}

	var activeRoom *models.RoomData
	if room, ok := Manager.ActiveRooms[uint(roomID)]; !ok {
		newRoom := &models.RoomData{
			Room: currentRoom,
			ActiveUsers: make(map[*models.UserData]bool),
			Broadcast: make(chan *models.Message, 100),
			Registered: make(chan *models.UserData, 100),
			Unregistered: make(chan *models.UserData, 100),
			Close: make(chan bool),
			TaskQueue: make(chan models.MessageTask, 1000),
		}
		activeRoom = newRoom
		Manager.AddRoom(newRoom)

		go newRoom.Run()
	} else {
		activeRoom = room
	}
	
	data := GetMessages(activeRoom.Room.ID)

	c.HTML(200, "chat.html", data)
}

func GetMessages(roomID uint) []models.Message {
	var messages []models.Message
	middlewares.DB.Where("room_id = ?", roomID).
		Order("created_at desc").
		Find(&messages)
	return messages
}

func ChatHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Не удалось установить соединение")
		c.JSON(400, gin.H{"Error": "Не удалось установить соединение с сокетом"})
		return
	}

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
	currentRoom.Registered <- currentUserData

	defer func() {
		currentRoom.Unregistered <- currentUserData
		if !currentRoom.CheckActive() {
			Manager.Mu.Lock()

			Manager.Delete(currentRoom)
			currentRoom.Stop()
			
			Manager.Mu.Unlock()
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
		currentRoom.Broadcast <- &message
	}
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