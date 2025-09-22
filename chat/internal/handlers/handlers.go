package handlers

import (
	"net/http"
	"strconv"

	"github.com/andro-kes/Chat/chat/internal/models"
	"github.com/andro-kes/Chat/chat/logger"
	"github.com/andro-kes/Chat/chat/responses"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin:     func(r *http.Request) bool { return true },
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Log.Error(
			"Не удалось обновить соединение websocket",
			zap.Error(err),
		)
		responses.SendJSONResponse(w, 500, map[string]any{
			"Error": "Не удалось обновить соединение websocket",
		})
	}
	logger.Log.Info("Соединение установлено")

	url := r.URL
	query := url.Query()
	id := query.Get("id")

	roomID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		logger.Log.Error(
			"Не удалось спарсить id комнаты",
			zap.Any("id", roomID),
			zap.Error(err),
		)
		responses.SendJSONResponse(w, 500, map[string]any{
			"Error": "Неверный идентификатор комнаты",
		})
		return
	}

	defer conn.Close()

	for {
		var message models.Message
		err := conn.ReadJSON(&message)
		if err != nil {
			logger.Log.Warn(
				"Не удалось считать сообщение",
				zap.Error(err),
			)
			responses.SendJSONResponse(w, 400, map[string]any{
				"Error": "Не удалось прочитать сообщение",
			})
			return
		}

	}
}


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


func CreateRoom(c *gin.Context) {
	currentUser := getCurrentUser(c)
	var roomName RoomName
	if err := c.ShouldBindBodyWithJSON(&roomName); err != nil {
		log.Println("Не удалось создать комнату")
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
	log.Println("Создана комната", NewRoom.Name)

	if err := middlewares.DB.Model(&NewRoom).Association("Users").Append(&currentUser); err != nil {
		log.Println("Ошибка добавления пользователя:", err)
		c.JSON(500, gin.H{"error": "Не удалось добавить пользователя в комнату"})
		return
	}

	c.JSON(200, gin.H{"CreateRoom": "Success", "room": NewRoom})
}


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


func MainPageHandler(c *gin.Context) {
	user := getCurrentUser(c)
	log.Println("Вход на главную страницу", user.Username)
	c.HTML(200, "main.html", gin.H{
		"UserName": user.Username,
	})
}