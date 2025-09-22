// ВРЕМЕННО: Пакет handlers содержит HTTP-хэндлеры аутентификации и управления
// пользователями. Включает OAuth (Яндекс) и стандартные операции: вход,
// регистрация, выход, обновление и установка пароля после OAuth. Логирование
// реализовано через zap. Комментарии временные для ориентира при рефакторинге.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/andro-kes/Chat/auth/binding"
	"github.com/andro-kes/Chat/auth/internal/models"
	"github.com/andro-kes/Chat/auth/internal/services"
	"github.com/andro-kes/Chat/auth/logger"
	"github.com/andro-kes/Chat/auth/responses"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/yandex"
)

// Управляющая структура, которая содержит все методы для работы с пользователем
type AuthHandlers struct {
	UserService services.UserService
}

func NewAuthHandlers() *AuthHandlers {
	return &AuthHandlers{
		UserService: services.NewUserService(),
	}
}

var oauth2Config *oauth2.Config

func initData() {
	oauth2Config = &oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8000/auth",
		Scopes:       []string{"login:info", "login:email"},
		Endpoint:     yandex.Endpoint,
	}
	
	if oauth2Config.ClientID == "" || oauth2Config.ClientSecret == "" {
		logger.Log.Fatal("Не обнаружены ClientID or ClientSecret")
	}
}

type YandexUser struct {
	Login        string   `json:"login"`
	DefaultEmail string   `json:"default_email"`
	Emails       []string `json:"emails"`
}

var (
	oauthStateString = os.Getenv("SECRET_KEY")
)

// Получает запрос пользователя на OAuth
// Перенаправляет запрос на LoginYandexHandler
func (*AuthHandlers) AuthYandexHandler(w http.ResponseWriter, r *http.Request) {
	initData()
	url := oauth2Config.AuthCodeURL(oauthStateString)
	logger.Log.Info(
		"Перенаправляю",
		zap.String("url", url),
	)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect) 
}

// LoginYandexHandler ВРЕМЕННО: обрабатывает редирект OAuth, проверяет state,
// обменивает код на токен, запрашивает профиль пользователя в Яндексе и
// передает данные в сервис пользователей. Возвращает JSON/HTML в зависимости
// от результата (создание пользователя или успешный вход).
func (ah *AuthHandlers) LoginYandexHandler(w http.ResponseWriter, r *http.Request) {
	initData()
	logger.Log.Info("LoginYandexHandler запущен")

	currentUrl := oauth2Config.AuthCodeURL(oauthStateString)
	query, err := url.Parse(currentUrl)
	if err != nil {
		logger.Log.Error(
			"Не удалось распарсить url",
			zap.Error(err),
		)
		responses.SendJSONResponse(w, 404, map[string]any{
			"Error": "не верный адрес страницы",
		})
		return
	}

	params := query.Query()

	state := params.Get("state")
	logger.Log.Info(
		"Получен state:", 
		zap.String("state", state),
	)
	if state != oauthStateString {
		logger.Log.Warn(
			fmt.Sprintf("Неверный статус, ожидалось: %s, получено: '%s'\n",
				oauthStateString, state,
			),
		)
		responses.SendJSONResponse(w, 400, map[string]any{
			"Error": "Неверный state",
		})
		return
	}

	code := params.Get("code")
	logger.Log.Info(
		"Получен код",
		zap.String("code", code),
	)

	token, err := oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		logger.Log.Error(
			"oauthConf.Exchange() не сработал",
			zap.Error(err),
		)
		responses.SendJSONResponse(w, 400, map[string]any{
			"Error": "oauthConf.Exchange() не сработал",
		})
		return
	}

	client := oauth2Config.Client(context.Background(), token)
	resp, err := client.Get("https://login.yandex.ru/info?format=json")
	if err != nil {
		logger.Log.Error(
			"Не удалось получить информацию о пользователе",
			zap.Error(err),
		)
		responses.SendJSONResponse(w, 500, map[string]any{
			"Error": "Не удалось получить информацию о пользователе",
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error(
			"Не удалось прочитать тело ответа",
			zap.Error(err),
		)
		responses.SendJSONResponse(w, 500, map[string]any{
			"Error": "Не удалось прочитать тело ответа",
		})
		return
	}

	var yandexUser YandexUser
	err = json.Unmarshal(body, &yandexUser)
	if err != nil {
		logger.Log.Error(
			"Не удалось разобрать JSON",
			zap.Error(err),
		)
		responses.SendJSONResponse(w, 500, map[string]any{
			"Error": "Не удалось разобрать JSON",
		})
		return
	}

	loginData, err := ah.UserService.OAuthLogin(yandexUser.Login, yandexUser.DefaultEmail)
	if err.Error() == "Пользователь был создан" {
		responses.SendHTMLResponse(w, 301, "auth.html", map[string]any{
			"title": "Добавить пароль",
			"username": loginData.User.Username,
			"email": loginData.User.Email,
		})
		return
	} else if err != nil {
		responses.SendJSONResponse(w, 400, map[string]any{
			"Error": "Не удалось войти с помощью OAuth",
		})
		return
	} else {
		responses.SendJSONResponse(w, 200, map[string]any{
			"Message": "Успешный вход",
			"AccessToken": loginData.AccessTokenString,
		})
		return
	}
}

// SetPassword ВРЕМЕННО: добавляет пароль пользователю после OAuth
func (au *AuthHandlers) SetPassword(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := binding.BindUserWithJSON(r, &user)
	if err != nil {
		logger.Log.Warn(
			"Невалидный пароль",
			zap.String("password", user.Password),
		)
		responses.SendJSONResponse(w, 400, map[string]any{
			"Error": "Невалидный пароль",
		})
		return
	}

	err = au.UserService.SetPassword(&user)
	if err != nil {
		responses.SendJSONResponse(w, 400, map[string]any{
			"Error": "Не удалось установить пароль",
		})
		return
	}
	responses.SendJSONResponse(w, 200, map[string]any{
		"Message": "Пароль установлен",
	})
}


// LoginPageHandler ВРЕМЕННО: отдает HTML-страницу входа
func (*AuthHandlers) LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	responses.SendHTMLResponse(w, 200, "login.html", map[string]any{
		"title": "login",
	})
}

// LoginHandler ВРЕМЕННО: аутентифицирует пользователя и устанавливает куки
// с refresh и access токенами с безопасными флагами
func (ah *AuthHandlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := binding.BindUserWithJSON(r, &user)
	if err != nil {
		responses.SendJSONResponse(w, 400, map[string]any{
			"Error": err.Error(),
		})
		return
	}

	loginData, err := ah.UserService.Login(&user)
	if err != nil {
		responses.SendJSONResponse(w, 400, map[string]any{
			"Error": "Не удалось войти в систему",
		})
		return
	}

	cookie := &http.Cookie{
        Name:     "refresh_token",
        Value:    loginData.RefreshTokenString,
        Path:     "/",
        HttpOnly: true, // Доступ только через HTTP, защита от XSS
        Secure:   true, // Только HTTPS
        SameSite: http.SameSiteStrictMode, // Защита от CSRF
    }
    http.SetCookie(w, cookie)

	cookie = &http.Cookie{
        Name:     "access_token",
        Value:    loginData.AccessTokenString,
        Path:     "/",
        HttpOnly: true, // Доступ только через HTTP, защита от XSS
        Secure:   true, // Только HTTPS
        SameSite: http.SameSiteStrictMode, // Защита от CSRF
    }
    http.SetCookie(w, cookie)

    responses.SendJSONResponse(w, 200, map[string]any{
        "Message": "Успешный вход в систему",
        "AccessToken": loginData.AccessTokenString,
        "RefreshToken": loginData.RefreshTokenString,
    })
}

// LogoutHandler ВРЕМЕННО: инвалидирует refresh токен и очищает куки
func (ah *AuthHandlers) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		logger.Log.Warn(
			"refresh token не найден",
		)
		responses.SendJSONResponse(w, 404, map[string]any{
			"Error": "Токен не найден",
		})
		return
	}

	err = ah.UserService.Logout(cookie.Value)
	if err != nil {
		responses.SendJSONResponse(w, 400, map[string]any{
			"Error": "Не удалось выйти из системы",
		})
		return
	}

	cookie = &http.Cookie{
        Name:   "refresh_token",
        Value:  "",
        Path:   "/",
        MaxAge: -1, // Удаление куки
    }
	http.SetCookie(w, cookie)

	cookie = &http.Cookie{
        Name:   "access_token",
        Value:  "",
        Path:   "/",
        MaxAge: -1, // Удаление куки
    }
	http.SetCookie(w, cookie)
	
	responses.SendJSONResponse(w, 200, map[string]any{
		"Message": "Пользователь вышел из системы",
	})
}

// SignUPPageHandler ВРЕМЕННО: отдает HTML-страницу регистрации
func (*AuthHandlers) SignUPPageHandler(w http.ResponseWriter, r *http.Request) {
	responses.SendHTMLResponse(w, 200, "signUp.html", map[string]any{
		"title": "sign_up_page",
	})
}

// SignUpHandler ВРЕМЕННО: регистрирует пользователя, хеширует пароль и
// устанавливает первичные куки с токенами
func (au *AuthHandlers) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := binding.BindUserWithJSON(r, &user); err != nil {
		logger.Log.Warn(
			"Не удалось получить данные пользователя",
			zap.Error(err),
		)
		responses.SendJSONResponse(w, 400, map[string]any{
			"Error": "Не удалось извлечь данные пользователя",
		})
	}

	loginData, err := au.UserService.SignUp(&user)
	if err != nil {
		responses.SendJSONResponse(w, 400, map[string]any{
			"Error": "Не удалось создать пользователя",
		})
	}

	cookie := &http.Cookie{
        Name:     "refresh_token",
        Value:    loginData.RefreshTokenString,
        Path:     "/",
        HttpOnly: true, // Доступ только через HTTP, защита от XSS
        Secure:   true, // Только HTTPS
        SameSite: http.SameSiteStrictMode, // Защита от CSRF
    }
    http.SetCookie(w, cookie)

	cookie = &http.Cookie{
        Name:     "access_token",
        Value:    loginData.AccessTokenString,
        Path:     "/",
        HttpOnly: true, // Доступ только через HTTP, защита от XSS
        Secure:   true, // Только HTTPS
        SameSite: http.SameSiteStrictMode, // Защита от CSRF
    }
    http.SetCookie(w, cookie)

    responses.SendJSONResponse(w, 200, map[string]any{
        "Message": "Новый пользователь успешно вошел в систему",
        "AccessToken": loginData.AccessTokenString,
        "RefreshToken": loginData.RefreshTokenString,
    })
}