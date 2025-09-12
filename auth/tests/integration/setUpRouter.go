package auth_tests

import (
	"log"

	"github.com/andro-kes/Chat/auth/internal/utils"
	


	"gorm.io/driver/sqlite"

)

func SetUpTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	
	DB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
    if err != nil {
        log.Fatal("Не удалось подключиться к SQLite")
    }

	DB.Migrator().DropTable(&models.User{})
	DB.Migrator().DropTable(&models.RefreshTokens{})
	DB.Migrator().DropTable(&models.Message{})
	DB.Migrator().DropTable(&models.Room{})

	DB.Migrator().CreateTable(&models.User{})
	DB.Migrator().CreateTable(&models.RefreshTokens{})
	DB.Migrator().CreateTable(&models.Message{})
	DB.Migrator().CreateTable(&models.Room{})

	hashPassword, err := utils.GenerateHashPassword("testpassword")
	if err != nil {
		log.Fatalln("Не удалось хэшировать тестовый пароль")
	}

	user := models.User{
		Email: "testemail",
		Password: string(hashPassword),
	}
	obj := DB.Create(&user)
	if obj.Error != nil {
		log.Fatalln("Не удалось создать тестового юзера")
	}

	utils.SetTestDB(DB)
	router.Use(func(c *gin.Context) {
        c.Set("DB", DB)
        c.Next()
    })

	return router
}