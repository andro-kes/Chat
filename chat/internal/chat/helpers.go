package chat

import (
	"log"

	"github.com/andro-kes/Chat/shared/models"
	"github.com/andro-kes/Chat/shared/middlewares"
)

func CheckAccess(room *models.Room, user *models.User) bool {
	var cnt int64
	middlewares.DB.Table("room_users").
		Where("room_id = ? AND user_id = ?", room.ID, user.ID).
		Count(&cnt)
	log.Println("Найдено совпадений", cnt)
	return cnt > 0
}