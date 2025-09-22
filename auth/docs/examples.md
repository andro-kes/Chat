# Примеры использования модуля аутентификации

## 🌐 Веб-интерфейс

### HTML формы

#### Страница входа (`login.html`)
```html
<!DOCTYPE html>
<html>
<head>
    <title>Вход в систему</title>
</head>
<body>
    <h1>Вход в систему</h1>
    <form id="loginForm">
        <div>
            <label for="email">Email:</label>
            <input type="email" id="email" name="email" required>
        </div>
        <div>
            <label for="password">Пароль:</label>
            <input type="password" id="password" name="password" required>
        </div>
        <button type="submit">Войти</button>
    </form>
    
    <div>
        <a href="/signup_page">Регистрация</a>
        <a href="/yandex_auth">Войти через Яндекс</a>
    </div>

    <script>
        document.getElementById('loginForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const formData = new FormData(e.target);
            const data = {
                email: formData.get('email'),
                password: formData.get('password')
            };
            
            try {
                const response = await fetch('/api/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    credentials: 'include',
                    body: JSON.stringify(data)
                });
                
                const result = await response.json();
                
                if (response.ok) {
                    alert('Успешный вход!');
                    // Перенаправление на главную страницу
                    window.location.href = '/dashboard';
                } else {
                    alert('Ошибка: ' + result.Error);
                }
            } catch (error) {
                alert('Ошибка сети: ' + error.message);
            }
        });
    </script>
</body>
</html>
```

#### Страница регистрации (`signUp.html`)
```html
<!DOCTYPE html>
<html>
<head>
    <title>Регистрация</title>
</head>
<body>
    <h1>Регистрация</h1>
    <form id="signupForm">
        <div>
            <label for="username">Имя пользователя:</label>
            <input type="text" id="username" name="username" required>
        </div>
        <div>
            <label for="email">Email:</label>
            <input type="email" id="email" name="email" required>
        </div>
        <div>
            <label for="password">Пароль:</label>
            <input type="password" id="password" name="password" required>
        </div>
        <button type="submit">Зарегистрироваться</button>
    </form>
    
    <div>
        <a href="/">Вход</a>
        <a href="/yandex_auth">Войти через Яндекс</a>
    </div>

    <script>
        document.getElementById('signupForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const formData = new FormData(e.target);
            const data = {
                username: formData.get('username'),
                email: formData.get('email'),
                password: formData.get('password')
            };
            
            try {
                const response = await fetch('/api/signup', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    credentials: 'include',
                    body: JSON.stringify(data)
                });
                
                const result = await response.json();
                
                if (response.ok) {
                    alert('Регистрация успешна!');
                    window.location.href = '/dashboard';
                } else {
                    alert('Ошибка: ' + result.Error);
                }
            } catch (error) {
                alert('Ошибка сети: ' + error.message);
            }
        });
    </script>
</body>
</html>
```

## 🔧 API клиенты

### JavaScript/TypeScript

#### Класс AuthClient
```typescript
class AuthClient {
    private baseUrl: string;
    
    constructor(baseUrl: string = 'http://localhost:8000') {
        this.baseUrl = baseUrl;
    }
    
    async login(email: string, password: string): Promise<LoginResponse> {
        const response = await fetch(`${this.baseUrl}/api/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'include',
            body: JSON.stringify({ email, password })
        });
        
        if (!response.ok) {
            throw new Error(`Login failed: ${response.status}`);
        }
        
        return response.json();
    }
    
    async signup(username: string, email: string, password: string): Promise<SignupResponse> {
        const response = await fetch(`${this.baseUrl}/api/signup`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'include',
            body: JSON.stringify({ username, email, password })
        });
        
        if (!response.ok) {
            throw new Error(`Signup failed: ${response.status}`);
        }
        
        return response.json();
    }
    
    async logout(): Promise<LogoutResponse> {
        const response = await fetch(`${this.baseUrl}/api/logout`, {
            method: 'POST',
            credentials: 'include'
        });
        
        if (!response.ok) {
            throw new Error(`Logout failed: ${response.status}`);
        }
        
        return response.json();
    }
    
    async initiateYandexAuth(): void {
        window.location.href = `${this.baseUrl}/yandex_auth`;
    }
}

// Использование
const authClient = new AuthClient();

// Вход
try {
    const result = await authClient.login('user@example.com', 'password123');
    console.log('Вход успешен:', result);
} catch (error) {
    console.error('Ошибка входа:', error);
}

// Регистрация
try {
    const result = await authClient.signup('newuser', 'newuser@example.com', 'password123');
    console.log('Регистрация успешна:', result);
} catch (error) {
    console.error('Ошибка регистрации:', error);
}
```

### Python

#### Класс AuthClient
```python
import requests
import json
from typing import Dict, Any

class AuthClient:
    def __init__(self, base_url: str = "http://localhost:8000"):
        self.base_url = base_url
        self.session = requests.Session()
    
    def login(self, email: str, password: str) -> Dict[str, Any]:
        """Вход в систему"""
        url = f"{self.base_url}/api/login"
        data = {"email": email, "password": password}
        
        response = self.session.post(url, json=data)
        response.raise_for_status()
        
        return response.json()
    
    def signup(self, username: str, email: str, password: str) -> Dict[str, Any]:
        """Регистрация пользователя"""
        url = f"{self.base_url}/api/signup"
        data = {"username": username, "email": email, "password": password}
        
        response = self.session.post(url, json=data)
        response.raise_for_status()
        
        return response.json()
    
    def logout(self) -> Dict[str, Any]:
        """Выход из системы"""
        url = f"{self.base_url}/api/logout"
        
        response = self.session.post(url)
        response.raise_for_status()
        
        return response.json()
    
    def get_yandex_auth_url(self) -> str:
        """Получить URL для Яндекс OAuth"""
        return f"{self.base_url}/yandex_auth"

# Использование
auth_client = AuthClient()

# Вход
try:
    result = auth_client.login("user@example.com", "password123")
    print("Вход успешен:", result)
except requests.exceptions.RequestException as e:
    print("Ошибка входа:", e)

# Регистрация
try:
    result = auth_client.signup("newuser", "newuser@example.com", "password123")
    print("Регистрация успешна:", result)
except requests.exceptions.RequestException as e:
    print("Ошибка регистрации:", e)
```

### Go

#### Клиент для тестирования
```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/http/cookiejar"
    "time"
)

type AuthClient struct {
    client  *http.Client
    baseURL string
}

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type SignupRequest struct {
    Username string `json:"username"`
    Email    string `json:"email"`
    Password string `json:"password"`
}

type AuthResponse struct {
    Message string `json:"Message"`
    User    *User  `json:"User,omitempty"`
    Error   string `json:"Error,omitempty"`
}

type User struct {
    ID        string    `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

func NewAuthClient(baseURL string) *AuthClient {
    jar, _ := cookiejar.New(nil)
    return &AuthClient{
        client: &http.Client{
            Jar:     jar,
            Timeout: 30 * time.Second,
        },
        baseURL: baseURL,
    }
}

func (c *AuthClient) Login(email, password string) (*AuthResponse, error) {
    req := LoginRequest{Email: email, Password: password}
    return c.post("/api/login", req)
}

func (c *AuthClient) Signup(username, email, password string) (*AuthResponse, error) {
    req := SignupRequest{Username: username, Email: email, Password: password}
    return c.post("/api/signup", req)
}

func (c *AuthClient) Logout() (*AuthResponse, error) {
    return c.post("/api/logout", nil)
}

func (c *AuthClient) post(endpoint string, data interface{}) (*AuthResponse, error) {
    var body io.Reader
    if data != nil {
        jsonData, err := json.Marshal(data)
        if err != nil {
            return nil, err
        }
        body = bytes.NewBuffer(jsonData)
    }

    req, err := http.NewRequest("POST", c.baseURL+endpoint, body)
    if err != nil {
        return nil, err
    }

    if data != nil {
        req.Header.Set("Content-Type", "application/json")
    }

    resp, err := c.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result AuthResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return &result, nil
}

// Пример использования
func main() {
    client := NewAuthClient("http://localhost:8000")

    // Регистрация
    signupResp, err := client.Signup("testuser", "test@example.com", "password123")
    if err != nil {
        fmt.Printf("Ошибка регистрации: %v\n", err)
        return
    }
    fmt.Printf("Регистрация: %s\n", signupResp.Message)

    // Вход
    loginResp, err := client.Login("test@example.com", "password123")
    if err != nil {
        fmt.Printf("Ошибка входа: %v\n", err)
        return
    }
    fmt.Printf("Вход: %s\n", loginResp.Message)

    // Выход
    logoutResp, err := client.Logout()
    if err != nil {
        fmt.Printf("Ошибка выхода: %v\n", err)
        return
    }
    fmt.Printf("Выход: %s\n", logoutResp.Message)
}
```

## 🧪 Тестирование

### Unit тесты

#### Тест UserService
```go
package services_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type MockUserRepo struct {
    mock.Mock
}

func (m *MockUserRepo) FindByEmail(email string) (*models.User, error) {
    args := m.Called(email)
    return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) CreateUser(user *models.User) error {
    args := m.Called(user)
    return args.Error(0)
}

func TestUserService_Login(t *testing.T) {
    // Arrange
    mockRepo := new(MockUserRepo)
    mockTokenService := new(MockTokenService)
    
    userService := &userService{
        Repo:         mockRepo,
        TokenService: mockTokenService,
    }
    
    testUser := &models.User{
        Email:    "test@example.com",
        Password: "password123",
    }
    
    existingUser := &models.User{
        ID:       uuid.New(),
        Email:    "test@example.com",
        Password: "$2a$16$hashedpassword", // bcrypt hash
    }
    
    mockRepo.On("FindByEmail", "test@example.com").Return(existingUser, nil)
    mockTokenService.On("GenerateRefreshToken", existingUser.ID).Return("refresh_token", nil)
    mockTokenService.On("GenerateAccessToken", existingUser.ID).Return("access_token", nil)
    
    // Act
    result, err := userService.Login(testUser)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, existingUser, result.User)
    assert.Equal(t, "refresh_token", result.RefreshTokenString)
    assert.Equal(t, "access_token", result.AccessTokenString)
    
    mockRepo.AssertExpectations(t)
    mockTokenService.AssertExpectations(t)
}
```

### Интеграционные тесты

#### Тест полного цикла аутентификации
```go
package integration_test

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "strings"
    "github.com/stretchr/testify/assert"
)

func TestAuthFlow(t *testing.T) {
    // Настройка тестовой базы данных
    setupTestDB(t)
    defer cleanupTestDB(t)
    
    // Создание тестового сервера
    server := httptest.NewServer(setupTestHandlers())
    defer server.Close()
    
    client := &http.Client{}
    
    t.Run("Полный цикл регистрации и входа", func(t *testing.T) {
        // Регистрация
        signupData := `{
            "username": "testuser",
            "email": "test@example.com",
            "password": "password123"
        }`
        
        signupReq, _ := http.NewRequest("POST", server.URL+"/api/signup", 
            strings.NewReader(signupData))
        signupReq.Header.Set("Content-Type", "application/json")
        
        signupResp, err := client.Do(signupReq)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusOK, signupResp.StatusCode)
        
        // Вход
        loginData := `{
            "email": "test@example.com",
            "password": "password123"
        }`
        
        loginReq, _ := http.NewRequest("POST", server.URL+"/api/login", 
            strings.NewReader(loginData))
        loginReq.Header.Set("Content-Type", "application/json")
        
        loginResp, err := client.Do(loginReq)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusOK, loginResp.StatusCode)
        
        // Выход
        logoutReq, _ := http.NewRequest("POST", server.URL+"/api/logout", nil)
        logoutResp, err := client.Do(logoutReq)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusOK, logoutResp.StatusCode)
    })
}
```

## 🔄 OAuth интеграция

### Настройка Яндекс OAuth

#### 1. Создание приложения
```bash
# Перейдите на https://oauth.yandex.ru/
# Создайте новое приложение
# Укажите redirect URI: http://localhost:8000/auth
# Выберите разрешения: login:info, login:email
```

#### 2. Тестирование OAuth flow
```javascript
// Инициация OAuth
function initiateYandexAuth() {
    window.location.href = 'http://localhost:8000/yandex_auth';
}

// Обработка callback (автоматически)
// После успешной авторизации пользователь будет перенаправлен на /auth
// с параметрами code и state
```

### Обработка OAuth в backend

```go
// Пример обработки OAuth callback
func (ah *AuthHandlers) LoginYandexHandler(w http.ResponseWriter, r *http.Request) {
    // Получение кода авторизации
    code := r.URL.Query().Get("code")
    state := r.URL.Query().Get("state")
    
    // Проверка state для защиты от CSRF
    if state != oauthStateString {
        http.Error(w, "Invalid state", http.StatusBadRequest)
        return
    }
    
    // Обмен кода на токен
    token, err := oauth2Config.Exchange(context.Background(), code)
    if err != nil {
        http.Error(w, "Token exchange failed", http.StatusBadRequest)
        return
    }
    
    // Получение данных пользователя
    client := oauth2Config.Client(context.Background(), token)
    resp, err := client.Get("https://login.yandex.ru/info?format=json")
    if err != nil {
        http.Error(w, "Failed to get user info", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()
    
    // Обработка данных пользователя...
}
```

## 📱 Мобильные приложения

### React Native

```javascript
import AsyncStorage from '@react-native-async-storage/async-storage';

class AuthService {
    constructor(baseURL = 'http://localhost:8000') {
        this.baseURL = baseURL;
    }
    
    async login(email, password) {
        try {
            const response = await fetch(`${this.baseURL}/api/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify({ email, password })
            });
            
            const data = await response.json();
            
            if (response.ok) {
                // Сохранение токенов
                await AsyncStorage.setItem('access_token', data.access_token);
                await AsyncStorage.setItem('refresh_token', data.refresh_token);
                return data;
            } else {
                throw new Error(data.Error);
            }
        } catch (error) {
            console.error('Login error:', error);
            throw error;
        }
    }
    
    async logout() {
        try {
            const refreshToken = await AsyncStorage.getItem('refresh_token');
            
            const response = await fetch(`${this.baseURL}/api/logout`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${refreshToken}`,
                },
            });
            
            // Очистка локального хранилища
            await AsyncStorage.multiRemove(['access_token', 'refresh_token']);
            
            return response.json();
        } catch (error) {
            console.error('Logout error:', error);
            throw error;
        }
    }
}

export default new AuthService();
```

## 🔧 Утилиты и хелперы

### Валидация данных

```javascript
// Валидация email
function isValidEmail(email) {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
}

// Валидация пароля
function isValidPassword(password) {
    return password.length >= 8 && 
           /[A-Z]/.test(password) && 
           /[a-z]/.test(password) && 
           /[0-9]/.test(password);
}

// Валидация имени пользователя
function isValidUsername(username) {
    return username.length >= 3 && 
           username.length <= 20 && 
           /^[a-zA-Z0-9_]+$/.test(username);
}
```

### Обработка ошибок

```javascript
class AuthError extends Error {
    constructor(message, statusCode, details = null) {
        super(message);
        this.name = 'AuthError';
        this.statusCode = statusCode;
        this.details = details;
    }
}

function handleAuthError(error) {
    if (error instanceof AuthError) {
        switch (error.statusCode) {
            case 400:
                return 'Неверные данные пользователя';
            case 401:
                return 'Неверный email или пароль';
            case 409:
                return 'Пользователь уже существует';
            default:
                return 'Произошла ошибка аутентификации';
        }
    }
    return 'Неизвестная ошибка';
}
```

Эти примеры помогут разработчикам быстро интегрировать модуль аутентификации в свои проекты и понять все возможности системы.



