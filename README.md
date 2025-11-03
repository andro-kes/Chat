# Chat — Микросервисный чат на Go

Проект "Chat" — учебный/демо комплект микросервисов для чата в реальном времени:
- auth — сервис аутентификации (HTTP + gRPC), выдаёт access/refresh JWT и хранит refresh-тokens в Postgres;
- chat — сервис чата (HTTP, WebSocket) с распределённой рассылкой сообщений через RabbitMQ и хранением сообщений в Postgres;
- web/templates — простые фронтенд-страницы для демонстрации.

README покрывает архитектуру, быстрый старт (локально и через Docker Compose), переменные окружения, API и рекомендации по безопасности и отладке.

Содержание
- Резюме
- Архитектура и компоненты
- Требования
- Переменные окружения
- Локальный запуск (Docker Compose)
- Запуск вручную (локальная разработка)
- Endpoints и API
- WebSocket
- gRPC (Auth)
- Схема БД и миграции
- RabbitMQ: гарантии доставки и параметры
- Безопасность и заметки по продакшн-конфигурации
- Отладка и типичные проблемы
- План работ / TODO для улучшения
- Контакты / как внести вклад

---

## Резюме
--------
Chat — это распределённый realtime-сервис. Клиенты подключаются по WebSocket к chat-сервису; сообщения публикуются в RabbitMQ и распределяются всем участникам комнаты. Сервис auth выдаёт и валидирует JWT (access + refresh), и предоставляет gRPC‑метод для других сервисов — GetUserId(token).

## Архитектура и компоненты
------------------------
- auth
  - HTTP endpoints: login, signup, OAuth (Яндекс), установка пароля.
  - gRPC server: AuthService.GetUserId(TokenRequest) -> UserIdResponse.
  - Хранение refresh tokens в Postgres.
- chat
  - HTTP + WebSocket сервер.
  - RabbitMQ publisher/consumer для распределения сообщений.
  - Postgres для хранения комнат и сообщений.
  - Middleware Auth вызывает auth gRPC для проверки access token.
- Общие:
  - Zap logger, шаблоны HTML в web/templates.
  - Переменные окружения для настройки подключений, секретов и путей.

## Требования
---------
- Go 1.24+ (рекомендовано)
- Postgres
- RabbitMQ (рекомендуется с management)
- (Опционально) Docker & Docker Compose для локального запуска

## Переменные окружения (основные)
--------------------------------
Ниже — переменные, которые используются сервисами. Обычно создают `.env` на основе `auth/env.example` / `chat/.env.example` (если есть).

Общие:
- SECRET_KEY — секрет для подписи JWT (обязательно, сильный секрет)
- TEMPLATE_DIR — путь к HTML шаблонам (опционально)

auth:
- DB_USER_URL — connection string для БД auth (например: `postgres://user:pass@postgres:5432/userdb?sslmode=disable`)
- CLIENT_ID, CLIENT_SECRET — для OAuth (Яндекс)
- AUTH_GRPC_ADDR — (опционально) адрес gRPC (для тестов)

chat:
- DB_CHAT_URL — connection string для БД chat (например: `postgres://user:pass@postgres:5432/chatdb?sslmode=disable`)
- AUTH_GRPC_ADDR — адрес auth gRPC (пример: `auth:50051`)
- RABBITMQ_USER, RABBITMQ_PASSWORD, RABBITMQ_ADDR — параметры подключения к RabbitMQ (addr в виде host:port)
- RABBITMQ_PREFETCH — prefetch (QoS) для консьюмера (рекомендуется 1)

## Запуск (рекомендуемый — Docker Compose)
--------------------------------------
Проект предназначен для запуска в локальной среде с помощью `docker compose`. В корне репозитория должен быть `docker-compose.yml` (если нет — создайте).

Пример сервисов в compose:
- postgres (для chat и auth, можно отдельные БД/контейнеры)
- rabbitmq (с management плагиом)
- auth (сборка Dockerfile в /auth)
- chat (сборка Dockerfile в /chat)

Пример (упрощённый):
1. Создайте `.env` с нужными значениями.
2. `docker compose up --build`
3. Посмотрите логи: `docker compose logs -f auth` / `chat`.
4. RabbitMQ Management — обычно: http://localhost:15672

## Запуск вручную (локальная разработка)
------------------------------------
1. Подготовьте окружение (Postgres, RabbitMQ).
2. Установите зависимости: `go mod download` внутри модулей.
3. Запустите auth:
   - `cd auth/cmd`
   - `go run .` (убедитесь, что окружение настроено: DB_USER_URL, SECRET_KEY, CLIENT_ID/SECRET)
4. Запустите chat:
   - `cd chat/cmd`
   - `go run .` (нужны DB_CHAT_URL, AUTH_GRPC_ADDR, RABBITMQ_* и SECRET_KEY)
5. Откройте http://localhost:8000 (auth) и http://localhost:8080 (chat) — шаблоны в web/templates.

## Endpoints / HTTP API
--------------------
Auth:
- GET  /               — Login page (HTML)
- GET  /yandex_auth    — redirect to Yandex OAuth
- GET  /auth           — OAuth callback
- POST /api/login      — JSON { "email", "password" } -> sets cookies (access_token, refresh_token) and returns tokens
- POST /api/signup     — JSON { "username", "email", "password" } -> creates user + sets cookies
- POST /api/logout     — invalidates refresh token (requires access cookies/Authorization)

Chat:
- GET  /               — main page (list rooms)
- POST /create         — create room (JSON: {"name": "..."})
- GET  /{id}           — room page (HTML)
- GET  /{id}/messages  — get messages (JSON) (handlers expect query param room_id in some endpoints — check code)
- WebSocket endpoint used in current code: /{id}/connect  (обратите внимание — frontend templates ожидают /{id}/ws; нужно согласовать путь; на момент анализа сервер регистрирует /{id}/connect)

## WebSocket
---------
- Сервер: в коде chat использует маршрут `/{id}/connect` (mux).
- Протокол: обмен JSON-пакетами. Клиент отправляет { "text": "<message>" } (frontend ожидает поле text).
- Рекомендации:
  - Использовать защищённые WebSocket (wss) в продакшн.
  - Проверять Origin и авторизовывать соединение (AuthMiddleware проверяет access token через gRPC).
  - На сервере генерируется объект Message {CreatedAt, SenderID, RoomID, Text/Content, ID} и публикуется в RabbitMQ; консьюмер рассылает структуру через WriteJSON.

## gRPC (Auth)
-----------
- Протобуф `auth/grpc/auth.proto`
- Сервис: `grpc.AuthService`
- Метод: `GetUserId(TokenRequest{token}) -> UserIdResponse{user_id}`
- chat.AuthMiddleware вызывает этот gRPC метод, чтобы получить user_id по bearer-токену.

## БД и миграции
-------------
Требуемые таблицы (примерная схема, адаптировать под миграции):
- users (id UUID PK, username, email unique, password hash, created_at, updated_at, deleted_at)
- refresh_tokens (user_id UUID, token_id UUID, token text, created_at) — индекс по user_id или token_id
- rooms (id UUID PK, name, users array/assoc, created_at, updated_at, deleted_at)
- room_users (room_id UUID, user_id UUID) — для быстрых JOIN'ов и доступа
- messages (id serial/UUID PK, room_id, user_id, content/text, created_at)

В коде database.Init вызывает `makeMigrations(ctx, pool)` — реализуйте миграции через golang-migrate / goose или SQL-скрипты. Перед запуском убедитесь, что миграции применены.

## RabbitMQ: ручной ack и гарантии доставки
---------------------------------------
Рекомендуется:
- Делать очередь durable.
- Публиковать с DeliveryMode = Persistent.
- Консьюмеры использовать `autoAck=false` и явно `delivery.Ack(false)` после успешной обработки.
- Настроить QoS (prefetch=1) для равномерной обработки.
- При ошибках использовать `Nack(requeue=true)` с разумной политикой retry и DLX (dead-letter) для сообщений, которые постоянно падают.

## Безопасность и продакшн-заметки
------------------------------
- SECRET_KEY должен быть сильным и храниться вне репозитория (Docker secrets, Vault).
- Cookies:
  - Secure=true в продакшн (HTTPS обязателен).
  - HttpOnly=true, SameSite=strict/none в зависимости от архитектуры.
- Не логировать токены и пароли.
- Снизьте bcrypt cost в dev, но в продакшн используйте 12–14 (в коде можно сделать BCRYPT_COST через env).
- Sanitize/escape сообщения перед вставкой в innerHTML на фронтенде — предотвратить XSS. Лучше отдавать текст в JSON и устанавливать textContent.
- WebSocket CheckOrigin должен быть ограничен списком доверенных origins.
- Rate limiting для login/signup endpoints.

## Отладка и типичные проблемы
---------------------------
- Ошибка: `nil pointer` в репозиториях — проверьте, вызван ли database.Init() до создания репозиториев.
- RabbitMQ: если consumer не получает сообщения — проверьте durable/queue name, management UI, prefetch.
- OAuth: callback handler должен брать `code` и `state` из `r.URL.Query()`, а не парсить AuthCodeURL.
- gRPC: chat должен подключаться к auth gRPC через grpc.Dial; если используется insecure (dev), укажите credentials/insecure.
- Cookie Secure=true не позволит установить куку по plain HTTP — в dev можно временно выставить Secure=false.

## Тесты и CI
----------
- Добавить unit-тесты для:
  - TokenService (generation/parsing/rotation).
  - RoomService.SendMessage (мок websocket.Conn).
  - Rabbit publish/consume (моки или интеграция с test Rabbit).
- Добавить интеграционные тесты под docker-compose (Postgres + RabbitMQ).
- В CI (GitHub Actions) — запуск `go test`, `go vet`, `golangci-lint`.

## План работ / TODO
-----------------
- Исправить frontend для работы с токенами.
- Согласовать WebSocket путь между frontend и server (`/{id}/ws` vs `/{id}/connect`).
- Добавить DLX/retry очередь для RabbitMQ.
- Сделать миграции в repo и интеграцию с database.Init (golang-migrate).
- Добавить health/readiness endpoints.
- Конвертировать hardcoded пути шаблонов в `TEMPLATE_DIR` через env.
- Улучшить обработку ошибок в handlers (всегда `return` после отправки ответа).
- Покрыть модуль тестами и добавить CI.

## Контрибьютинг
-------------
1. Форкните репозиторий.
2. Создайте ветку feature/bugfix.
3. Внесите изменения и откройте PR с описанием.
4. Добавляйте тесты и обновляйте README по мере необходимости.

Полезные ссылки
---------------
- RabbitMQ Management: http://localhost:15672
- PostgreSQL: pgAdmin / psql
- Go docs: https://pkg.go.dev

Автор / Контакты
----------------
Проект: andro-kes/Chat — учебный проект. Вопросы/PR приветствуются.