package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/andro-kes/Chat/auth/binding"
	"github.com/andro-kes/Chat/auth/internal/models"
	"github.com/andro-kes/Chat/auth/internal/services"
	"github.com/andro-kes/Chat/auth/logger"
	"github.com/andro-kes/Chat/auth/responses"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/yandex"
)

type authHandlers struct {
	UserService services.UserService
}

func NewAuthHandlers() *authHandlers {
	return &authHandlers{
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

func (*authHandlers) AuthYandexHandler(w http.ResponseWriter, r *http.Request) {
	initData()
	url := oauth2Config.AuthCodeURL(oauthStateString)
	logger.Log.Info(
		"Перенаправляю",
		zap.String("url", url),
	)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect) 
}

func (ah *authHandlers) LoginYandexHandler(w http.ResponseWriter) {
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

	ah.UserService.OAuthLogin(yandexUser.Login, yandexUser.DefaultEmail)
	var existingUser models.User
	DB.Where("email = ?", yandexUser.DefaultEmail).First(&existingUser)
	if existingUser.ID != 0 {
		log.Println("Вход в систему через Яндекс")
	} else {
		newUser := &models.User{
			Username: yandexUser.Login,
			Email:    yandexUser.DefaultEmail,
		}
		DB.Create(newUser)
		existingUser = *newUser
	}
	
	refreshToken, err := utils.GenerateRefreshToken(DB, existingUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tokenString, err := utils.GenerateAccessToken(existingUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка при создании токена"})
		return
	}
	expititionTime := time.Now().Add(5 * time.Minute)

	c.SetCookie("refresh_token", refreshToken, int(time.Now().Add(7*24*time.Hour).Unix()), "/", "localhost", false, true)
	c.SetCookie("token", tokenString, int(expititionTime.Unix()), "/", "localhost", false, true)

	c.HTML(http.StatusOK, "auth.html", gin.H{
		"message": "Успешная авторизация через Яндекс",
		"username": yandexUser.Login,
		"email":   yandexUser.DefaultEmail,
	})
}

func (*authHandlers) LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	responses.SendHTMLResponse(w, 200, "login.html", map[string]any{
		"title": "login",
	})
}

func (ah *authHandlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
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
		if err.Error() == "Пользователь был создан" {
			// Добавить handler для добавления пароля
			http.Redirect(w, r, "/set_password", http.StatusMovedPermanently)
		} else {
			responses.SendJSONResponse(w, 400, map[string]any{
				"Error": "Не удалось войти в систему",
			})
			return
		}
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
		"User": loginData.User,
	})
}

func (ah *authHandlers) LogoutHandler(w http.ResponseWriter, r *http.Request) {
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

	ah.UserService.Logout(cookie.Value)

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

func (*authHandlers) SignUPPageHandler(c *gin.Context) {
	c.HTML(200, "signUp.html", nil)
}

func (*authHandlers) SignUpHandler(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Println("Не удалось зарегистрировать пользователя")
		c.JSON(400, gin.H{
			"message": "Не удалось зарегистрировать пользователя",
			"error": err.Error(),
		})
	}

	DB := utils.GetDB(c)

	var existingUser models.User
	DB.Where("email = ?", user.Email).First(&existingUser)

	if existingUser.ID != 0 {
		c.JSON(400, gin.H{
			"message": "Пользователь с таким email уже существует",
		})
		return
	}

	password, err := utils.GenerateHashPassword(user.Password)
	if err != nil {
		c.JSON(400, gin.H{
			"message": "Ошибка хэширования пароля",
			"error": err.Error(),
		})
	}
	user.Password = string(password)

	DB.Create(&user)
	log.Println("Создан новый пользователь", user.Username)

	user.Password = ""

	refreshToken, err := utils.GenerateRefreshToken(DB, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tokenString, err := utils.GenerateAccessToken(existingUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка при создании токена"})
		return
	}
	expititionTime := time.Now().Add(5 * time.Minute)

	c.SetCookie("refresh_token", refreshToken, int(time.Now().Add(7*24*time.Hour).Unix()), "/", "localhost", false, true)
	c.SetCookie("token", tokenString, int(expititionTime.Unix()), "/", "localhost", false, true)

	c.JSON(200, gin.H{"message": "Создан новый пользователь"})
}

func (*authHandlers) UpdateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{
			"message": "Невалидные данные",
			"error": err.Error(),
		})
		return
	}

	DB := utils.GetDB(c)

	var existingUser models.User
	obj := DB.Where("email = ?", user.Email).First(&existingUser)
	if obj.Error != nil {
		c.JSON(400, gin.H{
			"message": "Пользователь не найден",
			"error": obj.Error.Error(),
		})
		return
	}

	existingUser.Username = user.Username

	password, err := utils.GenerateHashPassword(user.Password)
	if err != nil {
		c.JSON(400, gin.H{
			"message": "Не удалось получить hash пароля",
			"error": err.Error(),
		})
		return
	}
	existingUser.Password = string(password)

	obj = DB.Save(&existingUser)
	if obj.Error != nil {
		c.JSON(400, gin.H{
			"message": "Не удалось сохранить изменения",
			"error": obj.Error.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"message": "Данные успешно обновлены"})
}