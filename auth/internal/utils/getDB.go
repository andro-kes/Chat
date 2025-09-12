package utils

import (
	"log"

	

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