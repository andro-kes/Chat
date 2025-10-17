package handlers

import (
	"net/http"
	"strconv"

	"github.com/andro-kes/Chat/chat/binding"
	"github.com/andro-kes/Chat/chat/internal/models"
	"github.com/andro-kes/Chat/chat/internal/services"
	"github.com/andro-kes/Chat/chat/logger"
	"github.com/andro-kes/Chat/chat/responses"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type ChatHandlers struct {
	ChatService services.ChatService
}

// NewChatHandlers создает и возвращает новый экземпляр обработчика чата.
// Обработчик использует сервис для выполнения операций, связанных с чатом (например,
// отправка/прием сообщений, управление комнатами). Сервис инициализируется через
// `services.NewChatService()`.
// Пример использования:
//   handler := NewChatHandlers()
func NewChatHandlers() *ChatHandlers {
	return &ChatHandlers{
		ChatService: services.NewChatService(),
	}
}

func (*ChatHandlers) ChatHandler(w http.ResponseWriter, r *http.Request) {
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

/*
The best idea now is refactoring structure of chat
Can use structure with saving conns and ConnManager, RoomManager, ChatManager
Something like this:
- Пользователь подключается через WebSocket → создаётся Connection.
- Hub.Register <- conn — хаб регистрирует соединение.
- Пользователь отправляет сообщение → попадает в Hub.Messages.
- Хаб вызывает RoomManager.Broadcast(roomID, message).
- Room.Broadcast рассылает сообщение всем Connection.Send.
- Каждый Connection пишет в свой WebSocket.

Summary I have to build new architecture with middleware, channels and structs

And I offer to update docs and comments for i can get this shit in a week or even more
*/
func (ch *ChatHandlers) ChatPageHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL
	query := url.Query()
	id := query.Get("id")
	roomID, err := uuid.Parse(id)
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

	user := r.Context().Value("user_id")
	currentUserID, err := uuid.Parse(user.(string))
	if err != nil {
		logger.Log.Warn(
			"Некорректный id пользователя",
		)
		responses.SendJSONResponse(w, 400, map[string]any{
			"Error": "Invalid User Id",
		})
		return
	}

	// check access must be in ChatService
	if !ch.ChatService.CheckAccess(currentUserID) {
		logger.Log.Warn(
			"У пользователя нет доступа в эту комнату",
		)
		responses.SendJSONResponse(w, 400, map[string]any{
			"Error": "Access denied",
		})
		return
	}

	if !ch.ChatService.IsActive(roomID) {
		roomService := services.NewRoomService(roomID)
		ch.ChatService.AddRoom(roomService)
		roomService.StartWorkers()
	}

	responses.SendHTMLResponse(w, 200, "chat.html", nil)
}

type RoomName struct {
	roomName string
}
func (ch *ChatHandlers) CreateRoom(w http.ResponseWriter, r *http.Request) {
	// TODO Вынести в отдельную функцию при рефакторинге
	user := r.Context().Value("user_id")
	currentUserID, err := uuid.Parse(user.(string))
	if err != nil {
		logger.Log.Warn(
			"Некорректный id пользователя",
		)
		responses.SendJSONResponse(w, 400, map[string]any{
			"Error": "Invalid User Id",
		})
		return
	}

	var roomName RoomName
	if err := binding.BindWithJSON(r, &roomName); err != nil {
		logger.Log.Error("Не удалось декодировать данные")
		responses.SendJSONResponse(w, 404, map[string]any{
			"Error": "Невалидное название комнаты",
		})
	}

	err = ch.ChatService.CreateRoom(roomName.roomName, currentUserID)
	if err != nil {
		logger.Log.Warn(
			"Комната с таким названием уже существует",
			zap.String("room_name", roomName.roomName),
		)
		responses.SendJSONResponse(w, 400, map[string]any{
			"Error": "Комната с таким названием уже существует",
		})
	}

	responses.SendJSONResponse(w, 301, map[string]any{
		"Message": "Room was created successfully",
	})
}


func (ch *ChatHandlers) GetRoomMessages(w http.ResponseWriter, r *http.Request) {
	url := r.URL
	query := url.Query()
	id := query.Get("room_id")

	if id == "" {
		responses.SendJSONResponse(w, 404, map[string]any{
			"Error": "Невалидный id комнаты",
		})
		logger.Log.Warn(
			"Не удалось спарсить id комнаты",
		)
		return
	}
	
	roomID, err := uuid.Parse(id)
	if err != nil {
		responses.SendJSONResponse(w, 404, map[string]any{
			"Error": "Невалидный id комнаты",
		})
		logger.Log.Warn(
			"Не удалось спарсить id комнаты",
			zap.Any("id", roomID),
			zap.Error(err),
		)
	}

	if !ch.ChatService.IsActive(roomID) {
		responses.SendJSONResponse(w, 404, map[string]any{
			"Error": "Комната не активна",
		})
		logger.Log.Warn(
			"Комната не активна",
		)
		return
	}

	room, err := ch.ChatService.GetRoom(roomID)
	if err != nil {
		responses.SendJSONResponse(w, 404, map[string]any{
			"Error": "Комната не найдена",
		})
		logger.Log.Warn(
			"Комната не найдена",
		)
	}

	messages := room.GetMessages()

	responses.SendJSONResponse(w, 200, map[string]any{
		"Messages": messages,
	})
}


func (ch *ChatHandlers) GetUserRooms(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user_id")
	currentUserID, err := uuid.Parse(user.(string))
	if err != nil {
		logger.Log.Warn(
			"Некорректный id пользователя",
		)
		responses.SendJSONResponse(w, 400, map[string]any{
			"Error": "Invalid User Id",
		})
		return
	}
    
	rooms := ch.ChatService.GetUserRooms(currentUserID)

	responses.SendJSONResponse(w, 200, map[string]any{
		"rooms": rooms,
	})
}

// MainPageHandler обрабатывает запрос к главной странице.
// 
// Функция извлекает идентификатор пользователя из контекста запроса,
// проверяет его корректность и отправляет HTML-страницу с данными пользователя.
// 
// Параметры:
//   - w *http.ResponseWriter: Интерфейс для записи HTTP-ответа.
//   - r *http.Request: HTTP-запрос, содержащий контекст с идентификатором пользователя.
// 
// Возвращает:
//   - 200 OK: Если всё прошло успешно, отправляется HTML-страница с `user_id`.
//   - 400 Bad Request: Если идентификатор пользователя некорректен.
// Пример использования:
//   http.HandleFunc("/", MainPageHandler)
func (*ChatHandlers)  MainPageHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user_id")
	currentUserID, err := uuid.Parse(user.(string))
	if err != nil {
		logger.Log.Warn(
			"Некорректный id пользователя",
		)
		responses.SendJSONResponse(w, 400, map[string]any{
			"Error": "Invalid User Id",
		})
		return
	}

	responses.SendHTMLResponse(w, 200, "main.html", map[string]any{
		"user_id": currentUserID,
	})
}