# Chat (Go) — сервис чата

Описание
--------
Chat — сервис реального времени на Go с использованием WebSocket и RabbitMQ для распределённой рассылки сообщений. Использует Postgres для хранения данных (rooms, messages) и внешний Auth-сервис (gRPC) для проверки токенов.

Компоненты
- HTTP API + WebSocket сервер (порт 8080)
- RabbitMQ (очередь `chat`) — публикация и рассылка сообщений
- Postgres — хранение rooms, messages, room_users
- Auth (gRPC) — проверка токенов

Требования
---------
- Docker и Docker Compose / docker compose
- Go 1.20+ (для локальной сборки)
- Postgres
- RabbitMQ (management)
- Auth service (gRPC) доступный по адресу, указанному в ENV

Переменные окружения (обязательные)
- DB_CHAT_URL - Postgres connection string для chat БД (например: `postgres://user:pass@postgres:5432/dbname`)
- RABBITMQ_USER - RabbitMQ user
- RABBITMQ_PASSWORD - RabbitMQ password
- SECRET_KEY - общий секрет для сервиса (используется приложением)
- (опционально) AUTH_GRPC_ADDR - адрес auth gRPC сервиса (например `auth:50051`)

Файл .env.example
```bash
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=chatdb

DB_CHAT_URL=postgres://postgres:postgres@postgres:5432/chatdb

RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest

SECRET_KEY=replace_this_secret
AUTH_GRPC_ADDR=auth:50051
```

Запуск для разработки (docker-compose)
1. Создайте `.env` из `.env.example` и заполните значения.
2. Запустите:
```bash
docker compose -f compose.yml up --build
```
3. Сервисы:
- Chat: http://localhost:8080
- Auth: http://localhost:8000 (если поднят)
- RabbitMQ Management: http://localhost:15672 (user/pass из env)
- Postgres: 5432

Запуск локально (без Docker)
1. Установите Go и зависимости: `go mod download`
2. Инициализируйте окружение (см. .env).
3. Запустите DB и RabbitMQ локально (или в Docker).
4. Запустите:
```bash
cd chat/cmd
go run main.go
```

API / Endpoints (интерфейс)
- GET / — главная страница (main.html) — список комнат
- POST /create — создание комнаты (JSON: { "name": "room name" })
- GET /{roomId} — страница комнаты (chat.html)
- WebSocket: ws(s)://host/{roomId}/ws — WebSocket подключение для отправки/получения сообщений
- GET /api/room/{roomId}/messages — получение прошлых сообщений (JSON)
- GET /api/rooms — список комнат пользователя

Замечания по API:
- Endpoints и пути должны быть согласованы между main.go и frontend (templates JS).
- Аутентификация: ожидается Authorization header с токеном; middleware проверяет токен через gRPC Auth service.

Безопасность
- Не храните секреты в репозитории.
- Установите CSP и корректную проверку Origin для WebSocket.
- Escape контент при выводе в HTML (front-end должен использовать textContent).
- Для продакшн — используйте Docker secrets / Hashicorp Vault для хранения секретов.

Рекомендации по развитию
- Реализовать миграции (migrate, goose, golang-migrate).
- Покрыть unit и integration тестами (особенно RoomService, repository).
- Добавить CI (GitHub Actions) для build/test/lint.
- Со временем — сделать docker multi-stage build и уменьшить размер образа.