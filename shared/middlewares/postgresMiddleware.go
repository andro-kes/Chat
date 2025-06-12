package middlewares

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/andro-kes/Chat/shared/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var DB *gorm.DB

func init() { 
	err := godotenv.Load("./.env")
	if err != nil {
		log.Fatal("Init DB: Не удалось загрузить configs")
	}

	DSN := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
		os.Getenv("HOST"),
		os.Getenv("USERNAME"),
		os.Getenv("PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	
	DB = openDB(DSN)
	models.DB = DB
}

func openDB(DSN string) *gorm.DB {
	var err error
	DB, err = gorm.Open(postgres.Open(DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err) 
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Ошибка при получении *sql.DB: %v", err)
		return nil
	}

    sqlDB.SetMaxIdleConns(10)  
    sqlDB.SetMaxOpenConns(100)   
    sqlDB.SetConnMaxLifetime(time.Hour) 

	if err := DB.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("Ошибка миграции: %v", err)
		return nil 
	}
	if err := DB.AutoMigrate(&models.Room{}); err != nil {
		log.Fatalf("Ошибка миграции: %v", err)
		return nil
	}
	if err := DB.AutoMigrate(&models.Message{}); err != nil {
		log.Fatalf("Ошибка миграции: %v", err)
		return nil
	}
	if err := DB.AutoMigrate(&models.RefreshTokens{}); err != nil {
		log.Fatalf("Ошибка миграции: %v", err)
		return nil
	}
	log.Println("Успешное подключение к базе данных и миграция выполнены")
	return DB
}

func DBMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		if DB == nil {
			log.Println("DBMiddleWare: База данных не инициализирована")
			c.AbortWithStatusJSON(500, gin.H{"error": "База данных не инициализирована"})
			return
		}
		log.Println(gin.Mode())
		c.Set("DB", DB)
		c.Next()
	}
}