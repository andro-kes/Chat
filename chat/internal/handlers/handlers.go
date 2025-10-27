package handlers

import (
	"net/http"

	"github.com/andro-kes/Chat/chat/binding"
	"github.com/andro-kes/Chat/chat/internal/models"
	"github.com/andro-kes/Chat/chat/internal/rabbit"
	"github.com/andro-kes/Chat/chat/internal/services"
	"github.com/andro-kes/Chat/chat/logger"
	"github.com/andro-kes/Chat/chat/responses"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type ChatHandlers struct {
	ChatService services.ChatService
	RabbitManager rabbit.RabbitManager
}

// NewChatHandlers создает и возвращает новый экземпляр обработчика чата.
// Обработчик использует сервис для выполнения операций, связанных с чатом (например,
// отправка/прием сообщений, управление комнатами). Сервис инициализируется через
// `services.NewChatService()`.
// Пример использования:
//   handler := NewChatHandlers()
func NewChatHandlers() *ChatHandlers {
	rm, err := rabbit.Init()
	if err != nil {
		logger.Log.Fatal(
			"Не удалось инициализировать очередь сообщений",
			zap.String("error", err.Error()),
		)
	}
	return &ChatHandlers{
		ChatService: services.NewChatService(),
		RabbitManager: rm,
	}
}

// ChatHandler обрабатывает WebSocket-соединение для чата.
// 
// Функция:
// 1. Устанавливает соединение WebSocket.
// 2. Извлекает идентификатор комнаты из URL-запроса.
// 3. В цикле считывает сообщения от клиента и передает отправляет их в очередь.
// 
// Параметры:
//   - w *http.ResponseWriter: Интерфейс для записи HTTP-ответа.
//   - r *http.Request: HTTP-запрос, содержащий параметр `id` комнаты.
// 
// Возвращает:
//   - 500 Internal Server Error: При ошибке установки соединения.
//   - 400 Bad Request: При некорректном формате `id` комнаты.
// 
// Пример использования:
//   http.HandleFunc("/chat", ChatHandler)
func (ch *ChatHandlers) ChatHandler(w http.ResponseWriter, r *http.Request) {
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

	for {
		var message models.Message
		err := conn.ReadJSON(&message)
		if err != nil {
			logger.Log.Warn(
				"Не удалось считать сообщение",
				zap.String("error", err.Error()),
			)
			conn.Close()
			return
		}

		err = ch.RabbitManager.PublishMessage(message)
		if err != nil {
			logger.Log.Warn(
				"Не удалось добавить сообщение в очередь",
				zap.String("error", err.Error()),
			)
			conn.Close()
			return
		}
	}
}

// ChatPageHandler отдает HTML-страницу чата после проверки доступа.
// 
// Функция:
// 1. Извлекает идентификатор комнаты из URL-запроса.
// 2. Проверяет, имеет ли пользователь доступ к комнате.
// 3. Если комната не активна, запускает её.
// 
// Параметры:
//   - w *http.ResponseWriter: Интерфейс для записи HTTP-ответа.
//   - r *http.Request: HTTP-запрос, содержащий параметр `id` комнаты и контекст пользователя.
// 
// Возвращает:
//   - 200 OK: Если всё прошло успешно, отправляется HTML-страница.
//   - 400 Bad Request: При некорректном `id` пользователя или комнаты.
//   - 403 Forbidden: Если у пользователя нет доступа.
// 
// Пример использования:
//   http.HandleFunc("/chat_page", ChatPageHandler)
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

	currentUserID, err := getUser(r)
	if err != nil {
		responses.SendJSONResponse(w, 400, map[string]any{
			"error": "Не удалось получить данные о пользователе",
		})
	}

	if !ch.ChatService.CheckAccess(*currentUserID) {
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
	}

	responses.SendHTMLResponse(w, 200, "chat.html", nil)
}

// helper struct for parsing room name
type RoomName struct {
	Name string `json:"name"`
}

// CreateRoom создает новую комнату для текущего пользователя.
// 
// Функция:
// 1. Извлекает идентификатор пользователя из контекста.
// 2. Десериализует JSON-запрос с названием комнаты.
// 3. Создает комнату через сервис.
// 
// Параметры:
//   - w *http.ResponseWriter: Интерфейс для записи HTTP-ответа.
//   - r *http.Request: HTTP-запрос, содержащий данные комнаты и контекст пользователя.
// 
// Возвращает:
//   - 201 Created: Если комната создана.
//   - 400 Bad Request: При некорректном `id` пользователя.
//   - 409 Conflict: Если комната с таким названием уже существует.
// 
// Пример использования:
//   http.HandleFunc("/create_room", CreateRoom)
func (ch *ChatHandlers) CreateRoom(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := getUser(r)
	if err != nil {
		responses.SendJSONResponse(w, 400, map[string]any{
			"error": "Не удалось получить данные о пользователе",
		})
	}

	var roomName RoomName
	if err := binding.BindWithJSON(r, &roomName); err != nil {
		logger.Log.Error("Не удалось декодировать данные")
		responses.SendJSONResponse(w, 404, map[string]any{
			"Error": "Невалидное название комнаты",
		})
	}

	err = ch.ChatService.CreateRoom(roomName.Name, *currentUserID)
	if err != nil {
		logger.Log.Warn(
			"Комната с таким названием уже существует",
			zap.String("room_name", roomName.Name),
		)
		responses.SendJSONResponse(w, 400, map[string]any{
			"Error": "Комната с таким названием уже существует",
		})
	}

	responses.SendJSONResponse(w, 301, map[string]any{
		"Message": "Room was created successfully",
	})
}

// GetRoomMessages возвращает список сообщений комнаты.
// 
// Функция:
// 1. Извлекает идентификатор комнаты из URL-запроса.
// 2. Проверяет, активна ли комната.
// 3. Возвращает сообщения через сервис.
// 
// Параметры:
//   - w *http.ResponseWriter: Интерфейс для записи HTTP-ответа.
//   - r *http.Request: HTTP-запрос, содержащий параметр `room_id`.
// 
// Возвращает:
//   - 200 OK: Список сообщений.
//   - 404 Not Found: Если комната не найдена или не активна.
// 
// Пример использования:
//   http.HandleFunc("/get_messages", GetRoomMessages)
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

	messages, err := room.GetMessages()
	if err != nil {
		logger.Log.Warn(
			"Не удалось получить сообщения",
			zap.String("room_id", roomID.String()),
		)
		responses.SendJSONResponse(w, 404, map[string]any{
			"Messages": messages,
		})
	}

	responses.SendJSONResponse(w, 200, map[string]any{
		"Messages": messages,
	})
}

// GetUserRooms возвращает список комнат, к которым имеет доступ текущий пользователь.
//
// Извлекает идентификатор пользователя из контекста запроса и вызывает сервис.
// Отправляет JSON-ответ с массивом комнат.
//
// Параметры:
//   - w *http.ResponseWriter: Интерфейс для записи HTTP-ответа.
//   - r *http.Request: HTTP-запрос, содержащий контекст с идентификатором пользователя.
//
// Возвращает:
//   - 200 OK: Если всё прошло успешно, отправляется JSON-ответ с массивом комнат.
//   - 400 Bad Request: Если идентификатор пользователя некорректен.
//   - 404 Not Found: Если комната не найдена.
//
// Пример использования:
//   http.HandleFunc("/get_user_rooms", ch.GetUserRooms)
func (ch *ChatHandlers) GetUserRooms(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := getUser(r)
	if err != nil {
		responses.SendJSONResponse(w, 400, map[string]any{
			"error": "Не удалось получить данные о пользователе",
		})
	}
    
	rooms, err := ch.ChatService.GetUserRooms(*currentUserID)
	if err != nil {
		logger.Log.Warn(
			"Не удалось получить списко комнат",
			zap.String("user_id", currentUserID.String()),
			zap.String("error", err.Error()),
		)
		responses.SendJSONResponse(w, 404, map[string]any{
			"rooms": rooms,
		})
	}

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
func (*ChatHandlers) MainPageHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := getUser(r)
	if err != nil {
		responses.SendJSONResponse(w, 400, map[string]any{
			"error": "Не удалось получить данные о пользователе",
		})
	}

	responses.SendHTMLResponse(w, 200, "main.html", map[string]any{
		"user_id": *currentUserID,
	})
}

func getUser(r *http.Request) (*uuid.UUID, error) {
	user := r.Context().Value("user_id")
	currentUserID, err := uuid.Parse(user.(string))
	if err != nil {
		logger.Log.Warn(
			"Некорректный id пользователя",
		)
		return nil, err
	}

	return &currentUserID, nil
}