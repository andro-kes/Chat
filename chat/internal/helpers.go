package chat

import (
	"log"
    "net/http"
    "time"

	"github.com/andro-kes/Chat/shared/middlewares"
	"github.com/andro-kes/Chat/shared/models"
	"github.com/gin-gonic/gin"
)

func CheckAccess(room *models.Room, user *models.User) bool {
	var cnt int64
	middlewares.DB.Table("room_users").
		Where("room_id = ? AND user_id = ?", room.ID, user.ID).
		Count(&cnt)
	log.Println("Найдено совпадений", cnt)
	return cnt > 0
}

func getCurrentRoom(roomID string) (models.Room, error) {
	var currentRoom models.Room
	obj := middlewares.DB.Where("id = ?", roomID).First(&currentRoom)
	return currentRoom, obj.Error
}

func getCurrentUser(c *gin.Context) models.User {
	user, ok := c.Get("User")
	if !ok {
		log.Println("MainPage: контекст не содержит данных о пользователе")
		return models.User{}
	}

	currentUser, ok := user.(models.User)
	if !ok {
		return models.User{}
	}

	return currentUser
}

// setAuthCookies устанавливает HttpOnly куки для access и refresh токенов
func setAuthCookies(c *gin.Context, accessToken string, refreshToken string) {
    secure := c.Request.TLS != nil
    http.SetCookie(c.Writer, &http.Cookie{
        Name:     "access_token",
        Value:    accessToken,
        Path:     "/",
        HttpOnly: true,
        Secure:   secure,
        SameSite: http.SameSiteStrictMode,
        Expires:  time.Now().Add(15 * time.Minute),
    })

    http.SetCookie(c.Writer, &http.Cookie{
        Name:     "refresh_token",
        Value:    refreshToken,
        Path:     "/",
        HttpOnly: true,
        Secure:   secure,
        SameSite: http.SameSiteStrictMode,
        Expires:  time.Now().Add(7 * 24 * time.Hour),
    })
}