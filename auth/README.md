# Auth Service (auth)

Сервис аутентификации для проекта Chat. Предоставляет:
- HTTP эндпоинты для входа/регистрации/OAuth (Яндекс)
- gRPC API (AuthService) для проверки токенов (GetUserId)
- Управление access/refresh токенами (JWT)
- Хранение refresh токенов в Postgres

Требования
- Go 1.20+
- Postgres
- (dev) Docker, docker-compose

Переменные окружения (обязательные)
- DB_USER_URL — Postgres connection string (например `postgres://user:pass@postgres:5432/userdb`)
- CLIENT_ID — OAuth client id (Яндекс)
- CLIENT_SECRET — OAuth client secret
- SECRET_KEY — секрет для подписи JWT
- TEMPLATE_DIR (опционально) — путь к шаблонам (по умолчанию `/app/web/templates`)

Пример `.env`:
```bash
DB_USER_URL=postgres://postgres:postgres@postgres:5432/userdb
CLIENT_ID=your_yandex_client_id
CLIENT_SECRET=your_yandex_client_secret
SECRET_KEY=replace_this_secret
TEMPLATE_DIR=/app/web/templates
```

Запуск (локально)
1. Поднять Postgres (локально или в Docker).
2. Создать .env с переменными.
3. Сборка и запуск:
```bash
cd auth/cmd
go run main.go
```

Запуск (docker)
- Сделать Dockerfile (см. корневой compose.yml) и docker compose up.

Endpoints HTTP
- GET  / — страница логина
- GET  /yandex_auth — редирект на OAuth авторизацию
- GET  /auth — callback OAuth (обрабатывает code & state)
- POST /api/login — login (JSON: { "email": "", "password": "" })
- POST /api/signup — signup (JSON: { "username": "", "email": "", "password": "" })
- POST /api/signup/set_password — установить пароль после OAuth
- POST /api/logout — logout (удаляет refresh token cookie)

gRPC
- AuthService.GetUserId(TokenRequest{token}) => UserIdResponse{user_id}

Безопасность и рекомендации
- Не логировать токены и пароли.
- В production обязательно HTTPS — cookies с Secure=true.
- Настроить rate limiting на /api/login и /api/signup.
- Хранение секретов: Docker secrets / Hashicorp Vault.
- Пароли — bcrypt cost 12-14 (конфигурируемо).

Миграции БД
- Используйте golang-migrate или похожий инструмент для создания таблиц users и refresh_tokens.

Тесты
- Добавить unit-тесты для TokenService (генерация/парсинг), UserService, репозиториев.