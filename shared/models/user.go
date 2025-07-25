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
	Password string `json:"password"`
	Rooms []*Room `gorm:"many2many:room_users"`
}

type UserData struct {
	User User
	Conn *websocket.Conn
	Mu sync.Mutex
}

func (userData *UserData) CloseConn() {
	userData.Mu.Lock()
	defer userData.Mu.Unlock()
	
	if userData.Conn != nil {
		userData.Conn.Close()
		userData.Conn = nil
	}
}

type Claims struct {
	jwt.StandardClaims
}