package models

import (
	"sync"

	"gorm.io/gorm"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/websocket"
)

type User struct {
	gorm.Model
	Username string
	Email string
	Rooms []*Room `gorm:"many2many:room_users"`
}

type UserData struct {
	User User
	Conn *websocket.Conn
	Mu sync.Mutex
}

type Claims struct {
	jwt.StandardClaims
}