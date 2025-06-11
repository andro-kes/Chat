package models

import (
	"gorm.io/gorm"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/websocket"
)

type User struct {
	gorm.Model
	Username string
	Email string
	Password string
	Rooms []*Room `gorm:"many2many:room_users"`
	Conn *websocket.Conn
}

type Claims struct {
	Role string
	jwt.StandardClaims
}