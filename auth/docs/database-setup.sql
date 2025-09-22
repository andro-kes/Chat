-- SQL скрипты для настройки базы данных модуля аутентификации
-- Выполните эти команды для создания необходимых таблиц и индексов

-- ===========================================
-- СОЗДАНИЕ БАЗЫ ДАННЫХ
-- ===========================================

-- Создание базы данных (выполните от имени суперпользователя)
-- CREATE DATABASE chat_db;
-- CREATE USER chat_user WITH PASSWORD 'secure_password_123';
-- GRANT ALL PRIVILEGES ON DATABASE chat_db TO chat_user;

-- Подключение к базе данных
-- \c chat_db;

-- ===========================================
-- СОЗДАНИЕ ТАБЛИЦ
-- ===========================================

-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL
);

-- Таблица refresh токенов
CREATE TABLE IF NOT EXISTS refresh_tokens (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ===========================================
-- СОЗДАНИЕ ИНДЕКСОВ
-- ===========================================

-- Индекс для быстрого поиска пользователей по email
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Индекс для поиска токенов по пользователю
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- Индекс для поиска токенов по token_id
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_id ON refresh_tokens(token_id);

-- Индекс для поиска активных пользователей (без deleted_at)
CREATE INDEX IF NOT EXISTS idx_users_active ON users(id) WHERE deleted_at IS NULL;

-- ===========================================
-- ТРИГГЕРЫ
-- ===========================================

-- Функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггер для автоматического обновления updated_at в таблице users
CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- ===========================================
-- ПРАВА ДОСТУПА
-- ===========================================

-- Предоставление прав пользователю приложения
GRANT SELECT, INSERT, UPDATE, DELETE ON users TO chat_user;
GRANT SELECT, INSERT, DELETE ON refresh_tokens TO chat_user;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO chat_user;

-- ===========================================
-- ТЕСТОВЫЕ ДАННЫЕ (ОПЦИОНАЛЬНО)
-- ===========================================

-- Вставка тестового пользователя (пароль: password123)
-- INSERT INTO users (username, email, password) VALUES 
-- ('testuser', 'test@example.com', '$2a$16$hashedpasswordexample');

-- ===========================================
-- ПРОВЕРКА НАСТРОЙКИ
-- ===========================================

-- Проверка создания таблиц
SELECT table_name FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_name IN ('users', 'refresh_tokens');

-- Проверка индексов
SELECT indexname FROM pg_indexes 
WHERE tablename IN ('users', 'refresh_tokens');

-- Проверка прав доступа
SELECT grantee, privilege_type 
FROM information_schema.role_table_grants 
WHERE table_name IN ('users', 'refresh_tokens');



