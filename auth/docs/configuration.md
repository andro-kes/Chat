# Настройка и запуск модуля аутентификации

## 🚀 Быстрый старт

### Предварительные требования

- **Go 1.24+** - основной язык программирования
- **Docker** и **Docker Compose** - для контейнеризации
- **PostgreSQL** - база данных (через Docker)
- **Git** - для клонирования репозитория

### Установка зависимостей

#### macOS (Homebrew)
```bash
brew install go docker docker-compose postgresql
```

#### Ubuntu/Debian
```bash
sudo apt update
sudo apt install golang-go docker.io docker-compose postgresql-client
```

#### Windows
- Скачайте Go с [golang.org](https://golang.org/dl/)
- Установите Docker Desktop
- Установите Git

## ⚙️ Конфигурация

### 1. Переменные окружения

Создайте файл `.env` в корне проекта:

```bash
# База данных
POSTGRES_USER=chat_user
POSTGRES_PASSWORD=secure_password_123
POSTGRES_DB=chat_db
DB_USER_URL=postgres://chat_user:secure_password_123@postgres:5432/chat_db

# Яндекс OAuth (получите на https://oauth.yandex.ru/)
CLIENT_ID=your_yandex_client_id
CLIENT_SECRET=your_yandex_client_secret

# JWT секретный ключ (сгенерируйте случайную строку)
SECRET_KEY=your_super_secret_jwt_key_here_make_it_long_and_random
```

### 2. Настройка Яндекс OAuth

1. Перейдите на [Яндекс OAuth](https://oauth.yandex.ru/)
2. Создайте новое приложение
3. Укажите redirect URI: `http://localhost:8000/auth`
4. Выберите разрешения: `login:info`, `login:email`
5. Скопируйте `Client ID` и `Client Secret` в `.env`

### 3. Генерация секретного ключа

```bash
# Сгенерируйте случайный ключ
openssl rand -base64 32

# Или используйте Python
python3 -c "import secrets; print(secrets.token_urlsafe(32))"
```

## 🐳 Запуск через Docker

### 1. Клонирование репозитория
```bash
git clone <repository-url>
cd Chat
```

### 2. Настройка переменных окружения
```bash
cp .env.example .env
# Отредактируйте .env файл
```

### 3. Запуск всех сервисов
```bash
docker-compose up -d
```

### 4. Проверка статуса
```bash
docker-compose ps
```

### 5. Просмотр логов
```bash
# Все сервисы
docker-compose logs

# Только auth сервис
docker-compose logs auth

# Следить за логами в реальном времени
docker-compose logs -f auth
```

## 🖥 Локальный запуск

### 1. Установка зависимостей Go
```bash
cd auth
go mod download
```

### 2. Настройка базы данных

#### Запуск PostgreSQL через Docker
```bash
docker run -d \
  --name postgres-auth \
  -e POSTGRES_USER=chat_user \
  -e POSTGRES_PASSWORD=secure_password_123 \
  -e POSTGRES_DB=chat_db \
  -p 5432:5432 \
  postgres:latest
```

#### Создание таблиц
```sql
-- Подключитесь к базе данных
psql -h localhost -U chat_user -d chat_db

-- Создайте таблицы
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL
);

CREATE TABLE refresh_tokens (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Создайте индексы для производительности
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_id ON refresh_tokens(token_id);
```

### 3. Установка переменных окружения
```bash
export DB_USER_URL="postgres://chat_user:secure_password_123@localhost:5432/chat_db"
export CLIENT_ID="your_yandex_client_id"
export CLIENT_SECRET="your_yandex_client_secret"
export SECRET_KEY="your_super_secret_jwt_key_here"
```

### 4. Запуск приложения
```bash
go run cmd/main.go
```

## 🧪 Тестирование

### Запуск тестов
```bash
# Все тесты
go test ./...

# Интеграционные тесты
go test ./tests/integration/...

# Unit тесты
go test ./tests/unit/...

# С подробным выводом
go test -v ./...

# С покрытием кода
go test -cover ./...
```

### Настройка тестовой базы данных
```bash
# Создайте тестовую базу данных
createdb chat_db_test

# Установите переменную для тестов
export DB_USER_URL_TEST="postgres://chat_user:secure_password_123@localhost:5432/chat_db_test"
```

## 🔧 Разработка

### Структура проекта для разработки
```
auth/
├── cmd/                    # Точка входа
├── internal/              # Внутренняя логика
│   ├── handlers/          # HTTP обработчики
│   ├── services/          # Бизнес-логика
│   ├── repository/        # Доступ к данным
│   ├── models/            # Модели данных
│   ├── database/          # Конфигурация БД
│   └── utils/             # Утилиты
├── web/                   # Веб-интерфейс
├── docs/                  # Документация
├── tests/                 # Тесты
└── configs/               # Конфигурация
```

### Полезные команды для разработки

```bash
# Форматирование кода
go fmt ./...

# Проверка кода
go vet ./...

# Установка зависимостей
go mod tidy

# Обновление зависимостей
go get -u ./...

# Сборка приложения
go build -o auth ./cmd/main.go

# Запуск с горячей перезагрузкой (требует air)
air
```

### Настройка IDE

#### VS Code
Установите расширения:
- Go (Google)
- Go Test Explorer
- Docker

Настройки `.vscode/settings.json`:
```json
{
    "go.testFlags": ["-v"],
    "go.coverOnSave": true,
    "go.lintOnSave": "package",
    "go.vetOnSave": "package"
}
```

## 📊 Мониторинг и логи

### Просмотр логов

#### Docker
```bash
# Логи auth сервиса
docker-compose logs auth

# Логи с временными метками
docker-compose logs -t auth

# Последние 100 строк
docker-compose logs --tail=100 auth
```

#### Локальный запуск
Логи выводятся в stdout с использованием zap logger.

### Мониторинг базы данных
```bash
# Подключение к PostgreSQL
docker exec -it <postgres_container> psql -U chat_user -d chat_db

# Проверка соединений
SELECT * FROM pg_stat_activity;

# Размер базы данных
SELECT pg_size_pretty(pg_database_size('chat_db'));
```

## 🚨 Устранение неполадок

### Частые проблемы

#### 1. Ошибка подключения к базе данных
```
Error: failed to connect to database
```
**Решение**:
- Проверьте, что PostgreSQL запущен
- Убедитесь в правильности `DB_USER_URL`
- Проверьте сетевые настройки Docker

#### 2. Ошибка OAuth
```
Error: oauthConf.Exchange() не сработал
```
**Решение**:
- Проверьте `CLIENT_ID` и `CLIENT_SECRET`
- Убедитесь, что redirect URI настроен правильно
- Проверьте, что приложение активно в Яндекс OAuth

#### 3. Ошибка JWT токенов
```
Error: Не удалось сгенерировать refresh token
```
**Решение**:
- Проверьте `SECRET_KEY` - он не должен быть пустым
- Убедитесь, что ключ достаточно длинный (минимум 32 символа)

#### 4. Порт уже занят
```
Error: listen tcp :8000: bind: address already in use
```
**Решение**:
```bash
# Найдите процесс, использующий порт
lsof -i :8000

# Остановите процесс
kill -9 <PID>

# Или измените порт в коде
```

### Отладка

#### Включение debug логов
В `logger/logger.go`:
```go
config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
```

#### Проверка переменных окружения
```bash
# Показать все переменные
env | grep -E "(DB_|CLIENT_|SECRET_)"
```

## 🔒 Безопасность в production

### Рекомендации

1. **Используйте HTTPS**:
   ```bash
   # Настройте SSL сертификаты
   # Обновите redirect URI на https://
   ```

2. **Сильные пароли**:
   ```bash
   # Генерируйте случайные пароли для БД
   openssl rand -base64 32
   ```

3. **Ограничьте доступ к БД**:
   ```sql
   -- Создайте отдельного пользователя для приложения
   CREATE USER auth_app WITH PASSWORD 'strong_password';
   GRANT SELECT, INSERT, UPDATE, DELETE ON users TO auth_app;
   GRANT SELECT, INSERT, DELETE ON refresh_tokens TO auth_app;
   ```

4. **Регулярные обновления**:
   ```bash
   # Обновляйте зависимости
   go get -u ./...
   go mod tidy
   ```

## 📈 Производительность

### Оптимизация базы данных
```sql
-- Анализ производительности
EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'test@example.com';

-- Создание индексов
CREATE INDEX CONCURRENTLY idx_users_email_hash ON users USING hash(email);
```

### Мониторинг ресурсов
```bash
# Использование памяти Docker контейнерами
docker stats

# Логи производительности
docker-compose logs auth | grep -E "(slow|timeout|error)"
```
