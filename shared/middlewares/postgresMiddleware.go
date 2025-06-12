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
)

var DB *gorm.DB

func init() { 
	if os.Getenv("POSTGRES_HOST") == "" {
		log.Fatal("No Host")
	}
	if os.Getenv("POSTGRES_USER") == "" {
		log.Fatal("No USERNAME")
	}
	if os.Getenv("POSTGRES_PASSWORD") == "" {
		log.Fatal("No PASSWORD")
	}
	if os.Getenv("POSTGRES_DB") == "" {
		log.Fatal("No DB")
	}

	DSN := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
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