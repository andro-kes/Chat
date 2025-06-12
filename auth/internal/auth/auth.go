package auth

import (
	"time"
	"log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/yandex"
	"context"
	"io"
	"encoding/json"
	"os"

	"github.com/andro-kes/Chat/shared/models"
	"github.com/andro-kes/Chat/auth/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var oauth2Config *oauth2.Config

func init() {
	oauth2Config = &oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8000/auth",
		Scopes:       []string{"login:info", "login:email"}, 
		Endpoint:     yandex.Endpoint,
	}
}

type YandexUser struct {
	Login string `json:"login"`
	DefaultEmail string `json:"default_email"`
	Emails []string `json:"emails"`
}

var (
	oauthStateString = "random-string-for-state" // TODO Защита от CSRF
)

func AuthYandexHandler(c *gin.Context) {
	c.HTML(200, "index.html", nil)
}

func LoginYandexHandler(c *gin.Context) {
	state := c.Query("state")
	if state != oauthStateString {
		log.Printf("Неверный статус, ожидалось: '%s', получено: '%s'\n", oauthStateString, state)
		c.AbortWithStatus(401)
		return
	}

	code := c.Query("code")
	token, err := oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("oauthConf.Exchange() не сработал из-за: '%s'\n", err)
		c.AbortWithStatus(400)
		return
	}

	client := oauth2Config.Client(context.Background(), token)
	resp, err := client.Get("https://login.yandex.ru/info?format=json")
	if err != nil {
		log.Printf("Не удалось получить информацию о пользователе: %v", err)
		c.AbortWithStatus(500)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Не удалось прочитать тело ответа: %v", err)
		c.AbortWithStatus(500)
		return
	}

	var yandexUser YandexUser
	err = json.Unmarshal(body, &yandexUser)
	if err != nil {
		log.Printf("Не удалось разобрать JSON: %v", err)
		c.AbortWithStatus(500)
		return
	}

	db, ok := c.Get("DB")
	if !ok {
		c.JSON(400, gin.H{"Error": "Нет связи с базой данных"})
		return
	}
	DB, ok := db.(*gorm.DB)
	if !ok {
		c.JSON(400, gin.H{"Error": "Неверный формат DB"})
		return
	}

	var existingUser models.User
	DB.Where("email = ?", yandexUser.DefaultEmail).First(&existingUser)
	if existingUser.ID != 0 {
		log.Println("Вход в систему через Яндекс")
	} else {
		newUser := &models.User{
			Username: yandexUser.Login,
			Email: yandexUser.DefaultEmail,
		}
		DB.Create(newUser)
	}

	refreshToken, err := utils.GenerateRefreshToken(DB, existingUser.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	
	tokenString, err := utils.GenerateAccessToken(existingUser)
	if err != nil {
		c.JSON(400, gin.H{"error": "Ошибка при создании токена"})
		return
	}
	expititionTime := time.Now().Add(5 * time.Minute)

	c.SetCookie("refresh_token", refreshToken, int(time.Now().Add(7 * 24 * time.Hour).Unix()), "/", "localhost", false, true)
	c.SetCookie("token", tokenString, int(expititionTime.Unix()), "/", "localhost", false, true)
	c.JSON(200, gin.H{
		"message": "Успешная авторизация через Яндекс",
		"email":   yandexUser.DefaultEmail,
	})
}