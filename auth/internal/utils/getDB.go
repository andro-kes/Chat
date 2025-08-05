package utils

import (
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetDB(c *gin.Context) *gorm.DB {
	db, ok := c.Get("DB")
	if !ok {
		log.Fatalln("Не удалось получить доступ к базе")
	}
	DB, ok := db.(*gorm.DB)
	if !ok {
		log.Fatalln("Неверный формат типа для БД")
	}
	return DB
}