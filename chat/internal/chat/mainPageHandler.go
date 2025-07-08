package chat

import (
	"log"
	
	"github.com/gin-gonic/gin"
)

func MainPageHandler(c *gin.Context) {
	user := getCurrentUser(c)
	log.Println("Вход на главную страницу", user.Username)
	c.HTML(200, "main.html", gin.H{
		"UserName": user.Username,
	})
}