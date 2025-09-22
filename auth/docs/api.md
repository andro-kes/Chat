# API Документация - Модуль аутентификации

## Обзор

Модуль аутентификации предоставляет REST API для управления пользователями и сессиями. Все API endpoints работают на порту `8000` по умолчанию.

**Базовый URL**: `http://localhost:8000`

## 🔐 Аутентификация

### Типы токенов

- **Access Token** - короткоживущий токен (5 минут) для авторизации запросов
- **Refresh Token** - долгоживущий токен (24 часа) для обновления access токена

### Формат токенов

Токены передаются через HTTP куки:
- `access_token` - access токен
- `refresh_token` - refresh токен

## 📋 Endpoints

### 1. Страницы интерфейса

#### Главная страница входа
```http
GET /
```

**Описание**: Отображает HTML страницу входа

**Ответ**: HTML страница (`login.html`)

---

#### Страница регистрации
```http
GET /signup_page
```

**Описание**: Отображает HTML страницу регистрации

**Ответ**: HTML страница (`signUp.html`)

---

### 2. Аутентификация пользователей

#### Вход в систему
```http
POST /api/login
Content-Type: application/json
```

**Описание**: Аутентификация пользователя по email и паролю

**Тело запроса**:
```json
{
  "email": "user@example.com",
  "password": "userpassword"
}
```

**Успешный ответ** (200):
```json
{
  "Message": "Успешный вход в систему",
  "User": {
    "id": "uuid",
    "username": "username",
    "email": "user@example.com",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

**Ошибки**:
- `400` - Неверные данные пользователя
- `400` - Не удалось войти в систему

**Куки**: Устанавливаются `access_token` и `refresh_token`

---

#### Регистрация пользователя
```http
POST /api/signup
Content-Type: application/json
```

**Описание**: Создание нового пользователя

**Тело запроса**:
```json
{
  "username": "newuser",
  "email": "newuser@example.com",
  "password": "newpassword"
}
```

**Успешный ответ** (200):
```json
{
  "Message": "Новый пользователь успешно вошел в систему"
}
```

**Ошибки**:
- `400` - Не удалось извлечь данные пользователя
- `400` - Не удалось создать пользователя
- `400` - Пользователь с таким email уже существует

**Куки**: Устанавливаются `access_token` и `refresh_token`

---

#### Выход из системы
```http
POST /api/logout
```

**Описание**: Завершение сессии пользователя

**Куки**: Требуется `refresh_token`

**Успешный ответ** (200):
```json
{
  "Message": "Пользователь вышел из системы"
}
```

**Ошибки**:
- `404` - Токен не найден
- `400` - Не удалось выйти из системы

**Куки**: Очищаются `access_token` и `refresh_token`

---

### 3. OAuth аутентификация (Яндекс)

#### Инициация OAuth
```http
GET /yandex_auth
```

**Описание**: Перенаправляет пользователя на страницу авторизации Яндекс

**Ответ**: HTTP редирект на Яндекс OAuth

---

#### Обработка OAuth callback
```http
GET /auth
```

**Описание**: Обрабатывает редирект от Яндекс OAuth

**Параметры запроса**:
- `code` - код авторизации от Яндекс
- `state` - параметр состояния для защиты от CSRF

**Возможные ответы**:

**Новый пользователь** (301):
```html
<!-- HTML страница для установки пароля -->
```

**Существующий пользователь** (200):
```json
{
  "Message": "Успешный вход"
}
```

**Ошибки**:
- `400` - Неверный state
- `400` - oauthConf.Exchange() не сработал
- `500` - Не удалось получить информацию о пользователе
- `500` - Не удалось разобрать JSON
- `400` - Не удалось войти с помощью OAuth

---

## 🔒 Безопасность

### Защита куки

Все токены передаются через защищенные HTTP куки:

```go
cookie := &http.Cookie{
    Name:     "refresh_token",
    Value:    tokenString,
    Path:     "/",
    HttpOnly: true,    // Защита от XSS
    Secure:   true,    // Только HTTPS
    SameSite: http.SameSiteStrictMode, // Защита от CSRF
}
```

### Валидация данных

- Email должен быть валидным
- Пароль хешируется с помощью bcrypt (cost=16)
- JWT токены подписываются HMAC-SHA256
- OAuth state проверяется для защиты от CSRF

## 📊 Коды ответов

| Код | Описание |
|-----|----------|
| `200` | Успешный запрос |
| `301` | Редирект (OAuth новый пользователь) |
| `400` | Ошибка клиента (неверные данные) |
| `404` | Ресурс не найден |
| `500` | Внутренняя ошибка сервера |

## 🧪 Примеры использования

### JavaScript (fetch)

```javascript
// Вход в систему
const loginResponse = await fetch('/api/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  credentials: 'include', // Важно для куки
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'password123'
  })
});

const loginData = await loginResponse.json();
console.log(loginData);

// Выход из системы
const logoutResponse = await fetch('/api/logout', {
  method: 'POST',
  credentials: 'include'
});

const logoutData = await logoutResponse.json();
console.log(logoutData);
```

### cURL

```bash
# Вход в систему
curl -X POST http://localhost:8000/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}' \
  -c cookies.txt

# Выход из системы
curl -X POST http://localhost:8000/api/logout \
  -b cookies.txt
```

## 🔄 OAuth Flow

1. **Инициация**: `GET /yandex_auth`
2. **Редирект**: Пользователь перенаправляется на Яндекс
3. **Авторизация**: Пользователь авторизуется в Яндекс
4. **Callback**: Яндекс перенаправляет на `GET /auth?code=...&state=...`
5. **Обработка**: Сервер обменивает код на токен и получает данные пользователя
6. **Результат**: Создание нового пользователя или вход существующего

## 📝 Примечания

- Все временные комментарии в коде помечены как "ВРЕМЕННО"
- Модуль находится в активной разработке
- Для production использования рекомендуется настроить HTTPS
- Refresh токены автоматически инвалидируются при выходе



