package models

import (
	"sync"
	"log"
	"slices"
	"gorm.io/gorm"
)

var DB *gorm.DB

type Room struct {
	gorm.Model
	AdminID uint
	Admin User `gorm:"foreignKey:AdminID"`
	Users []*User `gorm:"many2many:room_users"`
	Messages []*Message `gorm:"foreignKey:RoomID"`
	Broadcast chan *Message
	Registered chan *User
	Unregistered chan *User
	Mu sync.Mutex
}

func (room *Room) Run() {
	for {
		select {
		case msg := <-room.Broadcast:
			room.Mu.Lock()
			msg.Room = room
			DB.Create(msg)
			msg.Save()
			room.Mu.Unlock()
		case user := <-room.Registered:
			ok := slices.Contains(room.Users, user)
			if !ok {
				room.Mu.Lock()
				err := DB.Model(&room).Association("Users").Append(user)
				if err != nil {
					log.Println("Models chat error: Не удалось добавить пользователя")
				}
				room.Mu.Unlock()
			}
		case user := <-room.Unregistered:
			user.Conn.Close()
		}
	}
}

type Message struct {
	gorm.Model
	SenderID uint
	Sender User `gorm:"foreignKey:SenderID"`
	RoomID uint
	Room *Room `gorm:"foreignKey:RoomID"`
	Text string
}

func (msg *Message) Save() {
	err := DB.Model(&msg.Room).Association("Messages").Append(msg)
	if err != nil {
		log.Println("Не удалось сохранить сообщение")
	}
}