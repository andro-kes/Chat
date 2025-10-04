# Диаграммы архитектуры модуля аутентификации

## 🏗 Общая архитектура

```mermaid
graph TB
    subgraph "HTTP Layer"
        H1[LoginHandler]
        H2[SignUpHandler]
        H3[LogoutHandler]
        H4[OAuthHandler]
    end
    
    subgraph "Business Layer"
        S1[UserService]
        S2[TokenService]
    end
    
    subgraph "Data Access Layer"
        R1[UserRepository]
        R2[TokenRepository]
    end
    
    subgraph "Database"
        DB[(PostgreSQL)]
        T1[users table]
        T2[refresh_tokens table]
    end
    
    subgraph "External Services"
        YA[Yandex OAuth]
    end
    
    H1 --> S1
    H2 --> S1
    H3 --> S1
    H4 --> S1
    H4 --> YA
    
    S1 --> S2
    S1 --> R1
    S2 --> R2
    
    R1 --> DB
    R2 --> DB
    
    DB --> T1
    DB --> T2
```

## 🔄 Процесс входа в систему

```mermaid
sequenceDiagram
    participant Client
    participant Handler
    participant UserService
    participant TokenService
    participant UserRepo
    participant Database

    Client->>Handler: POST /api/login
    Handler->>Handler: Validate input data
    Handler->>UserService: Login(user)
    UserService->>UserRepo: FindByEmail(email)
    UserRepo->>Database: SELECT user
    Database-->>UserRepo: User data
    UserRepo-->>UserService: User model
    UserService->>UserService: CompareHashPasswords()
    UserService->>TokenService: GenerateRefreshToken()
    TokenService-->>UserService: Refresh token
    UserService->>TokenService: GenerateAccessToken()
    TokenService-->>UserService: Access token
    UserService->>UserRepo: Save refresh token
    UserService-->>Handler: LoginData
    Handler->>Handler: Set HTTP cookies
    Handler-->>Client: JSON response + cookies
```

## 🔐 Процесс регистрации

```mermaid
sequenceDiagram
    participant Client
    participant Handler
    participant UserService
    participant TokenService
    participant UserRepo
    participant Utils
    participant Database

    Client->>Handler: POST /api/signup
    Handler->>Handler: Validate input data
    Handler->>UserService: SignUp(user)
    UserService->>UserRepo: FindByEmail(email)
    UserRepo->>Database: SELECT user
    Database-->>UserRepo: No user found
    UserRepo-->>UserService: Error (user not found)
    UserService->>Utils: GenerateHashPassword()
    Utils-->>UserService: Hashed password
    UserService->>UserRepo: CreateUser(user)
    UserRepo->>Database: INSERT user
    Database-->>UserRepo: User created
    UserRepo-->>UserService: Success
    UserService->>UserService: Login(user)
    UserService-->>Handler: LoginData
    Handler->>Handler: Set HTTP cookies
    Handler-->>Client: JSON response + cookies
```

## 🌐 OAuth процесс (Яндекс)

```mermaid
sequenceDiagram
    participant Client
    participant Handler
    participant Yandex
    participant UserService
    participant UserRepo
    participant Database

    Client->>Handler: GET /yandex_auth
    Handler->>Client: Redirect to Yandex
    Client->>Yandex: Authorization page
    Yandex->>Client: User authorization
    Client->>Yandex: Grant permissions
    Yandex->>Handler: GET /auth?code=...&state=...
    Handler->>Handler: Validate state
    Handler->>Yandex: Exchange code for token
    Yandex-->>Handler: Access token
    Handler->>Yandex: Get user info
    Yandex-->>Handler: User data
    Handler->>UserService: OAuthLogin(username, email)
    UserService->>UserRepo: FindByEmail(email)
    UserRepo->>Database: SELECT user
    
    alt User exists
        Database-->>UserRepo: User found
        UserRepo-->>UserService: User model
        UserService->>UserService: Login(user)
        UserService-->>Handler: LoginData
        Handler-->>Client: Success response
    else New user
        Database-->>UserRepo: No user found
        UserRepo-->>UserService: Error
        UserService->>UserRepo: CreateUser(user)
        UserRepo->>Database: INSERT user
        Database-->>UserRepo: User created
        UserRepo-->>UserService: Success
        UserService-->>Handler: "User created"
        Handler-->>Client: HTML form for password
    end
```

## 🚪 Процесс выхода из системы

```mermaid
sequenceDiagram
    participant Client
    participant Handler
    participant UserService
    participant TokenService
    participant TokenRepo
    participant Database

    Client->>Handler: POST /api/logout
    Handler->>Handler: Extract refresh token from cookie
    Handler->>UserService: Logout(token)
    UserService->>TokenService: ParseRefreshToken(token)
    TokenService-->>UserService: User ID
    UserService->>TokenService: RevokeRefreshToken(userID)
    TokenService->>TokenRepo: DeleteByID(userID)
    TokenRepo->>Database: DELETE refresh token
    Database-->>TokenRepo: Token deleted
    TokenRepo-->>TokenService: Success
    TokenService-->>UserService: Success
    UserService-->>Handler: Success
    Handler->>Handler: Clear HTTP cookies
    Handler-->>Client: Success response
```

## 🗄 Структура базы данных

```mermaid
erDiagram
    users {
        uuid id PK
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
        varchar username
        varchar email UK
        varchar password
    }
    
    refresh_tokens {
        uuid token_id PK
        uuid user_id FK
        text token
        timestamp created_at
    }
    
    users ||--o{ refresh_tokens : "has many"
```

## 🔧 Компоненты и зависимости

```mermaid
graph LR
    subgraph "Main Package"
        M[main.go]
    end
    
    subgraph "Handlers Package"
        AH[AuthHandlers]
    end
    
    subgraph "Services Package"
        US[UserService]
        TS[TokenService]
    end
    
    subgraph "Repository Package"
        UR[UserRepository]
        TR[TokenRepository]
    end
    
    subgraph "Models Package"
        U[User]
        RT[RefreshTokens]
    end
    
    subgraph "Utils Package"
        HP[HashPassword]
        CP[ComparePasswords]
    end
    
    subgraph "External Packages"
        JWT[jwt-go]
        BC[bcrypt]
        PGX[pgx]
        ZAP[zap]
    end
    
    M --> AH
    AH --> US
    US --> TS
    US --> UR
    TS --> TR
    UR --> U
    TR --> RT
    US --> HP
    US --> CP
    TS --> JWT
    HP --> BC
    UR --> PGX
    AH --> ZAP
```

## 🔒 Безопасность и токены

```mermaid
graph TB
    subgraph "Token Generation"
        TG[TokenService.GenerateRefreshToken]
        TA[TokenService.GenerateAccessToken]
    end
    
    subgraph "Token Storage"
        DB[(Database)]
        RT[refresh_tokens table]
    end
    
    subgraph "Token Validation"
        TV[TokenService.ParseRefreshToken]
        JV[JWT Validation]
    end
    
    subgraph "Security Features"
        BC[bcrypt Password Hashing]
        CS[CSRF Protection]
        XS[XSS Protection]
        SC[Secure Cookies]
    end
    
    TG --> RT
    TA --> JV
    RT --> DB
    TV --> JV
    BC --> CS
    CS --> XS
    XS --> SC
```

## 📊 Потоки данных

```mermaid
flowchart TD
    subgraph "Input Layer"
        HTTP[HTTP Request]
        JSON[JSON Data]
        FORM[Form Data]
    end
    
    subgraph "Processing Layer"
        VAL[Validation]
        BIZ[Business Logic]
        AUTH[Authentication]
    end
    
    subgraph "Storage Layer"
        REPO[Repository]
        DB[(Database)]
    end
    
    subgraph "Output Layer"
        RESP[HTTP Response]
        COOKIE[HTTP Cookies]
        HTML[HTML Template]
    end
    
    HTTP --> VAL
    JSON --> VAL
    FORM --> VAL
    VAL --> BIZ
    BIZ --> AUTH
    AUTH --> REPO
    REPO --> DB
    DB --> REPO
    REPO --> RESP
    RESP --> COOKIE
    RESP --> HTML
```

## 🧪 Тестирование

```mermaid
graph TB
    subgraph "Test Types"
        UT[Unit Tests]
        IT[Integration Tests]
        E2E[End-to-End Tests]
    end
    
    subgraph "Test Components"
        MS[Mock Services]
        MR[Mock Repositories]
        TC[Test Containers]
        TD[Test Database]
    end
    
    subgraph "Test Flow"
        SETUP[Setup Test Data]
        EXEC[Execute Test]
        ASSERT[Assert Results]
        CLEANUP[Cleanup]
    end
    
    UT --> MS
    IT --> MR
    E2E --> TC
    TC --> TD
    SETUP --> EXEC
    EXEC --> ASSERT
    ASSERT --> CLEANUP
```

## 🔄 Жизненный цикл токенов

```mermaid
stateDiagram-v2
    [*] --> TokenGeneration: User Login
    TokenGeneration --> AccessToken: Generate Access Token
    TokenGeneration --> RefreshToken: Generate Refresh Token
    AccessToken --> TokenValidation: API Request
    RefreshToken --> TokenStorage: Store in DB
    TokenValidation --> ValidToken: Token Valid
    TokenValidation --> InvalidToken: Token Invalid
    ValidToken --> AccessToken: Continue
    InvalidToken --> TokenRefresh: Use Refresh Token
    TokenRefresh --> TokenGeneration: Generate New Tokens
    TokenStorage --> TokenRevocation: User Logout
    TokenRevocation --> [*]: Session Ended
```

Эти диаграммы помогают визуализировать архитектуру и процессы модуля аутентификации, что облегчает понимание системы для новых разработчиков.







